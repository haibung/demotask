package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/demotask/backend/config"
	"github.com/demotask/backend/enum"
	"github.com/demotask/backend/models"
	billingPlan "github.com/demotask/backend/modules/billingPlan/repository"
	orderRepo "github.com/demotask/backend/modules/order/repository"
	paymentRepo "github.com/demotask/backend/modules/payment/repository"
	"github.com/demotask/backend/modules/product/repository"
	"github.com/demotask/backend/modules/subscription"
	subscriptionRepo "github.com/demotask/backend/modules/subscription/repository"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/paypal"
	"github.com/demotask/backend/utilities"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type (
	ISubscriptionController interface {
		CreateSubscription(ctx context.Context, req *subscription.CreateSubscriptionRequest, tx *gorm.DB) (*subscription.CreateSubscriptionResponse, error)
		// GetSubscription(ctx context.Context, id int) (*subscription.GetSubscriptionResponse, error)
		GetMySubscriptions(ctx context.Context, userID int) ([]subscription.GetSubscriptionResponse, error)
	}

	SubscriptionController struct {
		fx.In
		Config            *config.Config
		Logger            *logger.Logger
		SubscriptionRepo  subscriptionRepo.ISubscriptionRepository
		PaymentRepo       paymentRepo.IPaymentRepository
		ProductRepository repository.IProductRepository
		BillingPlanRepo   billingPlan.IBillingPlanRepository
		OrderRepo         orderRepo.IOrderRepository
		PayPalClient      paypal.IPayPalClient
	}
)

func NewController(controller SubscriptionController) ISubscriptionController {
	return &controller
}

func (c *SubscriptionController) GetMySubscriptions(ctx context.Context, userID int) ([]subscription.GetSubscriptionResponse, error) {

	subs, err := c.SubscriptionRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, utilities.ErrorRequest(err, http.StatusInternalServerError)
	}

	var responses []subscription.GetSubscriptionResponse

	for _, sub := range subs {
		responses = append(responses, subscription.GetSubscriptionResponse{
			ID:                   sub.ID,
			OrderID:              sub.OrderID,
			BillingPlanID:        sub.BillingPlanID,
			PaypalSubscriptionID: sub.PaypalSubscriptionID,
			Status:               sub.Status.String(),
			StartTime:            sub.StartTime,
			CreatedAt:            sub.CreatedAt,
		})
	}

	return responses, nil
}

func (c *SubscriptionController) CreateSubscription(ctx context.Context, req *subscription.CreateSubscriptionRequest, tx *gorm.DB) (*subscription.CreateSubscriptionResponse, error) {

	messages, err := req.Validate()
	if err != nil {
		return nil, utilities.ErrorRequest(err, http.StatusBadRequest, messages)
	}

	if req.ContextUserID == nil {
		return nil, utilities.ErrorRequest(errors.New("unauthorized"), http.StatusUnauthorized)
	}

	userID := *req.ContextUserID

	plan, err := c.BillingPlanRepo.FindByID(ctx, &models.BillingPlan{
		ID: req.BillingPlanID,
	})
	if err != nil {
		return nil, utilities.ErrorRequest(err, http.StatusNotFound)
	}

	if plan.PaypalPlanID == nil {
		return nil, utilities.ErrorRequest(errors.New("paypal_plan_id not found"), http.StatusBadRequest)
	}

	var regularCycle *models.BillingCycle
	for i := range plan.BillingCycles {
		if plan.BillingCycles[i].TenureType == "REGULAR" {
			regularCycle = &plan.BillingCycles[i]
			break
		}
	}

	if regularCycle == nil {
		return nil, utilities.ErrorRequest(errors.New("regular billing cycle not found"), http.StatusInternalServerError)
	}

	orderModel := &models.Order{
		UserID:    userID,
		Status:    enum.OrderStatusInternalPending,
		OrderType: enum.OrderTypeSubscription,
		Snapshot:  "{}",
	}

	orderID, err := c.OrderRepo.CreateOrder(ctx, orderModel, tx)
	if err != nil {
		return nil, utilities.ErrorRequest(err, http.StatusInternalServerError)
	}

	startTime := time.Now().UTC().Add(time.Minute)
	if req.StartTime != nil {
		startTime = req.StartTime.UTC()
	}

	ppReq := &paypal.CreateSubscriptionRequest{
		PlanID:    *plan.PaypalPlanID,
		Quantity:  req.Quantity,
		StartTime: utilities.StringPointer(startTime.Format(time.RFC3339)),
		ApplicationContext: &paypal.ApplicationContext{
			UserAction: "SUBSCRIBE_NOW",
			ReturnURL:  c.Config.PayPal.CheckoutSuccessURL,
			CancelURL:  c.Config.PayPal.CheckoutCancelURL,
		},
	}

	vaults, err := c.PaymentRepo.FindVaultByUserID(ctx, userID)
	useExistingVault := err == nil && len(vaults) > 0

	if useExistingVault {

		existingVault := vaults[0]

		ppReq.PaymentSource = &paypal.SubscriptionPaymentSource{
			PayPal: &paypal.SubscriptionPaypalSource{
				VaultID: existingVault.PayPalVaultID,
			},
		}

	} else {

		ppReq.PaymentSource = &paypal.SubscriptionPaymentSource{
			PayPal: &paypal.SubscriptionPaypalSource{
				Attributes: &paypal.SubscriptionPaymentAttributes{
					Vault: &paypal.SubscriptionVault{
						StoreInVault: "ON_SUCCESS",
						UsageType:    "MERCHANT",
					},
				},
			},
		}
	}
	ppRes, err := c.PayPalClient.CreateSubscription(ppReq)
	if err != nil {
		return nil, utilities.ErrorRequest(err, http.StatusBadGateway)
	}
	snapshot := models.SubscriptionSnapshot{
		Type:          "SUBSCRIPTION",
		BillingPlanID: plan.ID,
		PlanName:      plan.Name,
		PaypalPlanID:  *plan.PaypalPlanID,
		IntervalUnit:  regularCycle.IntervalUnit,
		IntervalCount: regularCycle.IntervalCount,
		Price:         regularCycle.PriceValue,
		Currency:      regularCycle.Currency,
		Quantity:      req.Quantity,
		TotalPrice:    regularCycle.PriceValue * float64(req.Quantity),
	}

	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		return nil, utilities.ErrorRequest(err, http.StatusInternalServerError)
	}

	sub := &models.Subscription{
		UserID:               userID,
		OrderID:              *orderID,
		BillingPlanID:        plan.ID,
		PaypalSubscriptionID: ppRes.ID,
		Status:               enum.SubscriptionStatusApprovalPending,
		StartTime:            &startTime,
		Snapshot:             string(snapshotJSON),
		Quantity:             req.Quantity,
	}

	if _, err := c.SubscriptionRepo.Create(ctx, sub, tx); err != nil {
		return nil, utilities.ErrorRequest(err, http.StatusInternalServerError)
	}

	if err := c.OrderRepo.UpdateStatus(
		ctx,
		*orderID,
		enum.OrderStatusInternalWaitingPayment,
		tx,
	); err != nil {
		return nil, utilities.ErrorRequest(err, http.StatusInternalServerError)
	}

	var approvalURL string
	for _, l := range ppRes.Links {
		if l.Rel == "approve" {
			approvalURL = l.Href
			break
		}
	}

	return &subscription.CreateSubscriptionResponse{
		OrderID:        *orderID,
		SubscriptionID: ppRes.ID,
		Status:         ppRes.Status,
		ApprovalURL:    approvalURL,
	}, nil
}
