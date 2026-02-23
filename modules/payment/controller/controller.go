package controller

import (
	"context"
	"errors"
	"net/http"

	"github.com/demotask/backend/enum"
	orderRepository "github.com/demotask/backend/modules/order/repository"
	"github.com/demotask/backend/modules/payment"
	"github.com/demotask/backend/modules/payment/repository"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/paypal"
	"github.com/demotask/backend/utilities"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type (
	IPaymentController interface {
		PayAgain(ctx context.Context, req *payment.PayAgainRequest, tx *gorm.DB) (*payment.PayAgainResponse, error)
		GetVaultTokens(ctx context.Context, req *payment.GetVaultTokensRequest) (*payment.GetVaultTokensResponse, error)
	}

	PaymentController struct {
		fx.In
		Logger       *logger.Logger
		PaymentRepo  repository.IPaymentRepository
		OrderRepo    orderRepository.IOrderRepository
		PayPalClient paypal.IPayPalClient
	}
)

func NewController(controller PaymentController) IPaymentController {
	return &controller
}

func (c *PaymentController) PayAgain(ctx context.Context, req *payment.PayAgainRequest, tx *gorm.DB) (*payment.PayAgainResponse, error) {

	messages, err := req.Validate()
	if err != nil {
		return nil, utilities.ErrorRequest(err, http.StatusBadRequest, messages)
	}

	orderModel, err := c.OrderRepo.FindByID(ctx, req.OrderID)
	if err != nil {
		return nil, utilities.ErrorRequest(err, http.StatusNotFound)
	}

	if orderModel.UserID != *req.ContextUserID {
		return nil, utilities.ErrorRequest(
			errors.New("forbidden"),
			http.StatusForbidden,
		)
	}

	if orderModel.Status != enum.OrderStatusInternalPending {
		return nil, utilities.ErrorRequest(
			errors.New("order not payable"),
			http.StatusConflict,
		)
	}

	if len(orderModel.Items) == 0 {
		return nil, utilities.ErrorRequest(
			errors.New("order has no items"),
			http.StatusBadRequest,
		)
	}

	vault, err := c.PaymentRepo.FindByID(ctx, req.VaultTokenID, *req.ContextUserID)
	if err != nil {
		return nil, utilities.ErrorRequest(err, http.StatusNotFound)
	}

	if vault.PayPalVaultID == "" {
		return nil, utilities.ErrorRequest(
			errors.New("invalid vault token"),
			http.StatusBadRequest,
		)
	}

	var totalAmount float64
	for _, item := range orderModel.Items {
		totalAmount += item.TotalPrice
	}

	if totalAmount <= 0 {
		return nil, utilities.ErrorRequest(
			errors.New("invalid order amount"),
			http.StatusBadRequest,
		)
	}

	currency := orderModel.Items[0].Currency

	createResp, err := c.PayPalClient.CreateOrderWithVault(totalAmount, currency, vault.PayPalVaultID)
	if err != nil {
		return nil, utilities.ErrorRequest(err, http.StatusBadGateway)
	}

	if err := c.OrderRepo.UpdatePaypalOrderID(ctx, orderModel.ID, createResp.ID, tx); err != nil {
		return nil, utilities.ErrorRequest(err, http.StatusInternalServerError)
	}

	if err := c.OrderRepo.UpdateStatus(ctx, orderModel.ID, enum.OrderStatusInternalWaitingPayment, tx); err != nil {
		return nil, utilities.ErrorRequest(err, http.StatusInternalServerError)
	}

	return &payment.PayAgainResponse{
		OrderID:       orderModel.ID,
		PaypalOrderID: createResp.ID,
		Status:        enum.OrderStatusInternalWaitingPayment.String(),
	}, nil
}

func (c *PaymentController) GetVaultTokens(ctx context.Context, req *payment.GetVaultTokensRequest) (*payment.GetVaultTokensResponse, error) {

	if req.ContextUserID == nil {
		return nil, utilities.ErrorRequest(errors.New("unauthorized"), http.StatusUnauthorized)
	}

	tokens, err := c.PaymentRepo.FindVaultByUserID(ctx, *req.ContextUserID)
	if err != nil {
		return nil, utilities.ErrorRequest(err, http.StatusInternalServerError)
	}

	var items []payment.VaultTokenItem
	for _, t := range tokens {
		items = append(items, payment.VaultTokenItem{
			ID:        t.ID,
			Email:     t.Email,
			IsDefault: t.IsDefault,
		})
	}

	return &payment.GetVaultTokensResponse{
		Items: items,
	}, nil
}
