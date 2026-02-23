package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/demotask/backend/enum"
	"github.com/demotask/backend/models"
	orderRepository "github.com/demotask/backend/modules/order/repository"
	paymentRepository "github.com/demotask/backend/modules/payment/repository"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/paypal"
	"github.com/demotask/backend/packages/postgres"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type (
	IPaymentCaptureHandler interface {
		HandleCapture(ctx context.Context, payload map[string]interface{}, tx *gorm.DB) error
		HandleVault(ctx context.Context, payload map[string]interface{}, tx *gorm.DB) error
		HandleOrderApproved(ctx context.Context, event paypal.WebhookEvent, tx *gorm.DB) error
	}

	PaymentCaptureHandler struct {
		fx.In
		DB           *postgres.DB
		Logger       *logger.Logger
		OrderRepo    orderRepository.IOrderRepository
		PaymentRepo  paymentRepository.IPaymentRepository
		PayPalClient paypal.IPayPalClient
	}
)

func NewPaymentCaptureHandler(h PaymentCaptureHandler) IPaymentCaptureHandler {
	return &h
}

func (h *PaymentCaptureHandler) HandleCapture(ctx context.Context, payload map[string]interface{}, tx *gorm.DB) error {

	resource, ok := payload["resource"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid payment capture payload")
	}

	supplementary, ok := resource["supplementary_data"].(map[string]interface{})
	if !ok {
		return nil
	}

	related, ok := supplementary["related_ids"].(map[string]interface{})
	if !ok {
		return nil
	}

	paypalOrderID, ok := related["order_id"].(string)
	if !ok || paypalOrderID == "" {
		return fmt.Errorf("missing paypal order id")
	}

	res := tx.WithContext(ctx).
		Model(&models.Order{}).
		Where("paypal_order_id = ?", paypalOrderID).
		Update("status", enum.OrderStatusInternalPaid)

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return fmt.Errorf("order not found: %s", paypalOrderID)
	}

	return nil
}

func (h *PaymentCaptureHandler) HandleVault(ctx context.Context, payload map[string]interface{}, tx *gorm.DB) error {

	resource, ok := payload["resource"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid vault payload: missing resource")
	}

	vaultID, ok := resource["id"].(string)
	if !ok || vaultID == "" {
		return fmt.Errorf("missing vault token id")
	}

	metadata, ok := resource["metadata"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing metadata in vault payload")
	}

	paypalOrderID, ok := metadata["order_id"].(string)
	if !ok || paypalOrderID == "" {
		return fmt.Errorf("missing order_id in metadata")
	}

	var order models.Order
	if err := tx.WithContext(ctx).
		Where("paypal_order_id = ?", paypalOrderID).
		First(&order).Error; err != nil {
		return fmt.Errorf("order not found for vault: %w", err)
	}

	var emailPtr *string
	if paymentSource, ok := resource["payment_source"].(map[string]interface{}); ok {
		if paypalObj, ok := paymentSource["paypal"].(map[string]interface{}); ok {
			if e, ok := paypalObj["email_address"].(string); ok && e != "" {
				emailPtr = &e
			}
		}
	}

	var existing models.VaultToken
	err := tx.WithContext(ctx).Where("paypal_vault_id = ?", vaultID).First(&existing).Error

	if err == nil {
		return nil
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if err := tx.WithContext(ctx).Model(&models.VaultToken{}).Where("user_id = ? AND is_default = ?", order.UserID, true).Update("is_default", false).Error; err != nil {
		return err
	}
	vault := models.VaultToken{
		UserID:        order.UserID,
		PayPalVaultID: vaultID,
		Email:         emailPtr,
		IsDefault:     true,
	}

	if err := tx.WithContext(ctx).Create(&vault).Error; err != nil {
		return err
	}

	h.Logger.Info(fmt.Sprintf("Vault token saved for user %d: %s", order.UserID, vaultID))

	return nil
}

func (c *PaymentCaptureHandler) HandleOrderApproved(ctx context.Context, event paypal.WebhookEvent, tx *gorm.DB) error {

	// 1. Parse resource safely
	resourceBytes, err := json.Marshal(event.Resource)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook resource: %w", err)
	}

	var orderResource paypal.CheckoutOrderResource
	if err := json.Unmarshal(resourceBytes, &orderResource); err != nil {
		return fmt.Errorf("failed to parse checkout order resource: %w", err)
	}

	if orderResource.ID == "" {
		return fmt.Errorf("paypal order id missing")
	}

	paypalOrderID := orderResource.ID

	// 2. Find local order
	ord, err := c.OrderRepo.FindOrderPaypalID(ctx, paypalOrderID)
	if err != nil {
		return err
	}

	if ord.Status != enum.OrderStatusInternalWaitingPayment {
		return nil
	}

	// 3. Capture order
	captureResp, err := c.PayPalClient.CaptureOrder(paypalOrderID)
	if err != nil {
		return err
	}

	if captureResp.Status != "COMPLETED" {
		return fmt.Errorf("capture not completed")
	}

	if len(captureResp.PurchaseUnits) == 0 ||
		len(captureResp.PurchaseUnits[0].Payments.Captures) == 0 {
		return fmt.Errorf("invalid capture response")
	}

	capture := captureResp.PurchaseUnits[0].Payments.Captures[0]

	amountFloat, err := strconv.ParseFloat(capture.Amount.Value, 64)
	if err != nil {
		return err
	}

	now := time.Now()
	rawBytes, _ := json.Marshal(captureResp)

	// 4. Idempotency check
	existing, _ := c.PaymentRepo.FindByCaptureID(ctx, capture.ID)
	if existing != nil {
		return nil
	}

	paymentModel := &models.Payment{
		OrderID:         &ord.ID,
		UserID:          ord.UserID,
		Provider:        "paypal",
		PaypalOrderID:   &paypalOrderID,
		PaypalCaptureID: &capture.ID,
		Amount:          amountFloat,
		Currency:        capture.Amount.CurrencyCode,
		PaidAt:          &now,
		Status:          enum.PaymentStatusCaptured,
		RawResponse:     string(rawBytes),
	}

	if _, err := c.PaymentRepo.Create(ctx, paymentModel, tx); err != nil {
		return err
	}

	return c.OrderRepo.UpdateStatus(
		ctx,
		ord.ID,
		enum.OrderStatusInternalPaid,
		tx,
	)
}
