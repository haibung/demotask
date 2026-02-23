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
	paymentRepository "github.com/demotask/backend/modules/payment/repository"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/paypal"
	"github.com/demotask/backend/packages/postgres"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type (
	ISubscriptionHandler interface {
		HandleSubscription(ctx context.Context, eventType string, payload map[string]interface{}, tx *gorm.DB) error
		HandleSale(ctx context.Context, payload map[string]interface{}, tx *gorm.DB) error
	}

	SubscriptionHandler struct {
		fx.In
		DB           *postgres.DB
		Logger       *logger.Logger
		PayPalClient paypal.IPayPalClient
		PaymentRepo  paymentRepository.IPaymentRepository
	}
)

func NewSubscriptionHandler(h SubscriptionHandler) ISubscriptionHandler {
	return &h
}

func (h *SubscriptionHandler) HandleSubscription(ctx context.Context, eventType string, payload map[string]interface{}, tx *gorm.DB) error {

	resource, ok := payload["resource"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid webhook payload: missing resource")
	}

	subscriptionID, ok := resource["id"].(string)
	if !ok || subscriptionID == "" {
		return fmt.Errorf("missing paypal subscription id")
	}

	var status enum.SubscriptionStatus
	updates := map[string]interface{}{}

	switch eventType {

	case "BILLING.SUBSCRIPTION.CREATED":
		status = enum.SubscriptionStatusApprovalPending

	case "BILLING.SUBSCRIPTION.ACTIVATED",
		"BILLING.SUBSCRIPTION.ACTIVE":

		status = enum.SubscriptionStatusActive

		if startTimeStr, ok := resource["start_time"].(string); ok {
			if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
				updates["start_time"] = t
			}
		}

	case "BILLING.SUBSCRIPTION.CANCELLED":
		status = enum.SubscriptionStatusCancelled

	case "BILLING.SUBSCRIPTION.SUSPENDED":
		status = enum.SubscriptionStatusSuspended

	default:
		h.Logger.Info("Unhandled subscription webhook event: " + eventType)
		return nil
	}

	var existing models.Subscription
	if err := tx.WithContext(ctx).
		Where("paypal_subscription_id = ?", subscriptionID).
		First(&existing).Error; err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	if existing.Status != status {
		updates["status"] = status
	}

	if err := tx.WithContext(ctx).
		Model(&models.Subscription{}).
		Where("paypal_subscription_id = ?", subscriptionID).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// if status == enum.SubscriptionStatusActive {
	// if err := tx.WithContext(ctx).
	// 	Model(&models.Order{}).
	// 	Where("id = ?", existing.OrderID).
	// 	Update("status", enum.OrderStatusInternalPaid).
	// 	Error; err != nil {
	// 	return fmt.Errorf("failed to update order: %w", err)
	// }

	if status == enum.SubscriptionStatusActive {

		if err := tx.WithContext(ctx).
			Model(&models.Order{}).
			Where("id = ?", existing.OrderID).
			Update("status", enum.OrderStatusInternalPaid).
			Error; err != nil {
			return fmt.Errorf("failed to update order: %w", err)
		}

		detail, err := h.PayPalClient.GetSubscriptionDetail(subscriptionID)
		if err != nil {
			h.Logger.Error("failed get subscription detail: " + err.Error())
			return nil
		}

		if detail.BillingInfo != nil {

			updates := map[string]interface{}{}

			if detail.BillingInfo.NextBillingTime != "" {
				if t, err := time.Parse(time.RFC3339,
					detail.BillingInfo.NextBillingTime); err == nil {
					updates["next_billing_time"] = t
				}
			}

			updates["failed_payments_count"] =
				detail.BillingInfo.FailedPaymentsCount

			if detail.BillingInfo.LastPayment != nil &&
				detail.BillingInfo.LastPayment.Time != "" {

				if t, err := time.Parse(time.RFC3339,
					detail.BillingInfo.LastPayment.Time); err == nil {
					updates["last_payment_time"] = t
				}
			}

			if len(updates) > 0 {
				tx.WithContext(ctx).
					Model(&models.Subscription{}).
					Where("id = ?", existing.ID).
					Updates(updates)
			}
		}

		if detail.PaymentSource != nil &&
			detail.PaymentSource.PayPal != nil {

			vaultID := detail.PaymentSource.PayPal.VaultID

			if vaultID != "" {

				var existingVault models.VaultToken
				err := tx.WithContext(ctx).
					Where("paypal_vault_id = ?", vaultID).
					First(&existingVault).Error

				if errors.Is(err, gorm.ErrRecordNotFound) {

					tx.WithContext(ctx).
						Model(&models.VaultToken{}).
						Where("user_id = ?", existing.UserID).
						Update("is_default", false)

					newVault := models.VaultToken{
						UserID:        existing.UserID,
						PayPalVaultID: vaultID,
						IsDefault:     true,
					}

					if err := tx.WithContext(ctx).
						Create(&newVault).Error; err == nil {

						tx.WithContext(ctx).
							Model(&models.Subscription{}).
							Where("id = ?", existing.ID).
							Update("vault_token_id", newVault.ID)
					}
				}
			}
		}
	}

	h.Logger.Info(
		fmt.Sprintf(
			"Subscription %s updated to status %s",
			subscriptionID,
			status.String(),
		),
	)

	return nil
}

func (h *SubscriptionHandler) HandleSale(ctx context.Context, payload map[string]interface{}, tx *gorm.DB) error {

	resource, ok := payload["resource"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid webhook payload: missing resource")
	}

	// 1. Extract subscription ID
	subscriptionID, ok := resource["billing_agreement_id"].(string)
	if !ok || subscriptionID == "" {
		return fmt.Errorf("missing billing_agreement_id")
	}

	// 2. Extract sale ID (unique transaction id)
	saleID, ok := resource["id"].(string)
	if !ok || saleID == "" {
		return fmt.Errorf("missing sale id")
	}

	// 3. Idempotency check
	existingPayment, _ := h.PaymentRepo.FindByCaptureID(ctx, saleID)
	if existingPayment != nil {
		return nil
	}

	// 4. Extract amount
	amountObj, ok := resource["amount"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing amount object")
	}

	valueStr, ok := amountObj["total"].(string)
	if !ok {
		return fmt.Errorf("missing amount total")
	}

	currency, ok := amountObj["currency"].(string)
	if !ok {
		return fmt.Errorf("missing currency")
	}

	amount, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return fmt.Errorf("invalid amount format: %w", err)
	}

	// 5. Extract PayPal payment time (do not use server time)
	createTimeStr, ok := resource["create_time"].(string)
	if !ok || createTimeStr == "" {
		return fmt.Errorf("missing create_time")
	}

	paidAt, err := time.Parse(time.RFC3339, createTimeStr)
	if err != nil {
		return fmt.Errorf("invalid create_time format: %w", err)
	}

	// 6. Find local subscription
	var subscription models.Subscription
	if err := tx.WithContext(ctx).
		Where("paypal_subscription_id = ?", subscriptionID).
		First(&subscription).Error; err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	rawBytes, _ := json.Marshal(payload)

	// 7. Create payment record
	payment := models.Payment{
		UserID:               subscription.UserID,
		SubscriptionID:       &subscription.ID,
		PaypalSubscriptionID: &subscriptionID,
		PaypalCaptureID:      &saleID,
		Provider:             "paypal",
		Amount:               amount,
		Currency:             currency,
		Status:               enum.PaymentStatusCaptured,
		PaidAt:               &paidAt,
		RawResponse:          string(rawBytes),
	}

	if err := tx.WithContext(ctx).Create(&payment).Error; err != nil {
		return err
	}

	detail, err := h.PayPalClient.GetSubscriptionDetail(subscriptionID)
	if err != nil {
		h.Logger.Error("failed get subscription detail: " + err.Error())
		return nil
	}

	if detail.BillingInfo != nil {

		updates := map[string]interface{}{}

		if detail.BillingInfo.NextBillingTime != "" {
			if t, err := time.Parse(time.RFC3339,
				detail.BillingInfo.NextBillingTime); err == nil {
				updates["next_billing_time"] = t
			}
		}

		updates["failed_payments_count"] =
			detail.BillingInfo.FailedPaymentsCount

		if detail.BillingInfo.LastPayment != nil &&
			detail.BillingInfo.LastPayment.Time != "" {

			if t, err := time.Parse(time.RFC3339,
				detail.BillingInfo.LastPayment.Time); err == nil {
				updates["last_payment_time"] = t
			}
		}

		if len(updates) > 0 {
			tx.WithContext(ctx).
				Model(&models.Subscription{}).
				Where("id = ?", subscription.ID).
				Updates(updates)
		}
	}

	return nil
}
