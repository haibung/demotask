package subscription

import (
	"errors"
	"time"

	"github.com/demotask/backend/packages/validation"
	"github.com/demotask/backend/utilities"
)

type (
	CreateSubscriptionRequest struct {
		ContextUserID *int
		BillingPlanID int        `json:"billing_plan_id"`
		VaultTokenID  *int       `json:"vault_token_id,omitempty"`
		StartTime     *time.Time `json:"start_time,omitempty"`
		Quantity      int        `json:"quantity"`

		Validation validation.Validation `json:"-"`
	}

	CreateSubscriptionResponse struct {
		OrderID        int    `json:"order_id"`
		SubscriptionID string `json:"subscription_id"`
		Status         string `json:"status"`
		ApprovalURL    string `json:"approval_url"`
	}

	GetSubscriptionRequest struct {
		ContextUserID *int
	}

	GetSubscriptionResponse struct {
		ID                   int        `json:"id"`
		OrderID              int        `json:"order_id"`
		BillingPlanID        int        `json:"billing_plan_id"`
		PaypalSubscriptionID string     `json:"paypal_subscription_id"`
		Status               string     `json:"status"`
		StartTime            *time.Time `json:"start_time"`
		CreatedAt            time.Time  `json:"created_at"`
	}
)

func (r *CreateSubscriptionRequest) Validate() ([]map[string]interface{}, error) {

	if r.ContextUserID == nil {
		r.Validation.Messages = append(r.Validation.Messages, map[string]interface{}{
			"user": "unauthorized",
		})
	}

	if r.BillingPlanID <= 0 {
		r.Validation.Messages = append(r.Validation.Messages, map[string]interface{}{
			"billing_plan_id": "billing_plan_id is required",
		})
	}

	if r.Quantity <= 0 {
		r.Validation.Messages = append(r.Validation.Messages, map[string]interface{}{
			"quantity": "quantity must be greater than 0",
		})
	}

	if len(r.Validation.Messages) > 0 {
		return r.Validation.Messages, errors.New(utilities.BadRequest)
	}

	return nil, nil
}
