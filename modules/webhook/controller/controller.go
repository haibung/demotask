package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/demotask/backend/enum"
	"github.com/demotask/backend/models"
	"github.com/demotask/backend/modules/webhook/handlers"
	"github.com/demotask/backend/modules/webhook/repository"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/paypal"
	"github.com/demotask/backend/utilities"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type IWebhookController interface {
	HandlePayPalWebhook(ctx context.Context, headers map[string]string, rawBody []byte, body map[string]interface{}, tx *gorm.DB) error
}

type WebhookController struct {
	fx.In
	Logger              *logger.Logger
	WebhookRepo         repository.IWebhookRepository
	PayPalClient        paypal.IPayPalClient
	SubscriptionHandler handlers.ISubscriptionHandler
	PaymentHandler      handlers.IPaymentCaptureHandler
}

func NewController(c WebhookController) IWebhookController {
	return &c
}

func (c *WebhookController) HandlePayPalWebhook(ctx context.Context, headers map[string]string, rawBody []byte, body map[string]interface{}, tx *gorm.DB) error {

	// 1. Extract event id
	eventID, ok := body["id"].(string)
	if !ok || eventID == "" {
		return utilities.ErrorRequest(
			fmt.Errorf("missing webhook event id"),
			http.StatusBadRequest,
		)
	}

	// 2. Extract event type
	eventType, ok := body["event_type"].(string)
	if !ok || eventType == "" {
		return utilities.ErrorRequest(
			fmt.Errorf("missing webhook event type"),
			http.StatusBadRequest,
		)
	}

	c.Logger.Info("Webhook received: " + eventType)

	// 3. Idempotency check
	existing, err := c.WebhookRepo.FindByEventID(ctx, eventID)
	if err == nil && existing != nil {
		c.Logger.Info("Duplicate webhook ignored: " + eventID)
		return nil
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return utilities.ErrorRequest(err, http.StatusInternalServerError)
	}

	// 4. Verify signature
	verificationReq := &paypal.WebhookVerificationRequest{
		AuthAlgo:         headers["paypal-auth-algo"],
		CertURL:          headers["paypal-cert-url"],
		TransmissionID:   headers["paypal-transmission-id"],
		TransmissionSig:  headers["paypal-transmission-sig"],
		TransmissionTime: headers["paypal-transmission-time"],
		WebhookID:        c.PayPalClient.GetWebhookID(),
		WebhookEvent:     rawBody,
	}

	valid, err := c.PayPalClient.VerifyWebhookSignature(verificationReq)
	if err != nil || !valid {
		c.Logger.Error("Invalid webhook signature: " + eventID)
		return utilities.ErrorRequest(
			fmt.Errorf("invalid webhook signature"),
			http.StatusUnauthorized,
		)
	}

	// 5. Save webhook event (pending)
	eventModel := &models.WebhookEvent{
		Provider:  "paypal",
		EventID:   eventID,
		EventType: eventType,
		Payload:   string(rawBody),
		Status:    enum.WebhookStatusPending,
	}

	eventIDDB, err := c.WebhookRepo.Save(ctx, eventModel, tx)
	if err != nil {
		return utilities.ErrorRequest(err, http.StatusInternalServerError)
	}

	// 6. Parse strongly typed event
	var parsedEvent paypal.WebhookEvent
	if err := json.Unmarshal(rawBody, &parsedEvent); err != nil {
		_ = c.WebhookRepo.MarkFailed(ctx, *eventIDDB, tx)
		return nil
	}

	// 7. Route event
	switch {

	case strings.HasPrefix(eventType, "BILLING.SUBSCRIPTION."):
		err = c.SubscriptionHandler.HandleSubscription(ctx, eventType, body, tx)

	case eventType == "CHECKOUT.ORDER.APPROVED":
		err = c.PaymentHandler.HandleOrderApproved(ctx, parsedEvent, tx)

	case eventType == "PAYMENT.CAPTURE.COMPLETED":
		err = c.PaymentHandler.HandleCapture(ctx, body, tx)

	case eventType == "PAYMENT.SALE.COMPLETED":
		err = c.SubscriptionHandler.HandleSale(ctx, body, tx)

	case eventType == "VAULT.PAYMENT-TOKEN.CREATED":
		err = c.PaymentHandler.HandleVault(ctx, body, tx)

	default:
		c.Logger.Info("Unhandled webhook event: " + eventType)
	}

	// 8. Handle result
	if err != nil {
		c.Logger.Error(err)
		_ = c.WebhookRepo.MarkFailed(ctx, *eventIDDB, tx)
		return nil
	}

	_ = c.WebhookRepo.MarkProcessed(ctx, *eventIDDB, tx)
	return nil
}
