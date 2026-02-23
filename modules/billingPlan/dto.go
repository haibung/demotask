package billingPlan

import (
	"errors"

	"github.com/demotask/backend/enum"
	"github.com/demotask/backend/packages/paypal"
	"github.com/demotask/backend/packages/validation"
	"github.com/demotask/backend/utilities"
)

type (
	CreateBillingPlanRequest struct {
		ContextUserID      *int
		ProductID          *int                        `json:"product_id"`
		Name               string                      `json:"name"`
		Description        string                      `json:"description"`
		Status             enum.PlanStatus             `json:"status"`
		BillingCycles      []CreateBillingCycleRequest `json:"billing_cycles"`
		PaymentPreferences *paypal.PaymentPreferences  `json:"payment_preferences"`
		Taxes              *CreatePlanTaxRequest       `json:"taxes"`

		Validation validation.Validation `json:"-"`
	}

	CreatePlanTaxRequest struct {
		Percentage float64 `json:"percentage"`
		Inclusive  bool    `json:"inclusive"`
	}

	CreateBillingCycleRequest struct {
		IntervalUnit  string  `json:"interval_unit"`
		IntervalCount int     `json:"interval_count"`
		TenureType    string  `json:"tenure_type"`
		Sequence      int     `json:"sequence"`
		TotalCycles   int     `json:"total_cycles"`
		Price         float64 `json:"price"`
		Currency      string  `json:"currency"`
	}

	CreateBillingPlanResponse struct {
		ID           int
		ProductID    int
		Name         string
		Status       string
		Description  string
		PaypalPlanID string
	}
)

func (r *CreateBillingPlanRequest) Validate() ([]map[string]interface{}, error) {
	r.Validation.IsEmptyString(r.Name, "name")

	if r.ProductID == nil {
		r.Validation.Messages = append(r.Validation.Messages, map[string]interface{}{
			"field": "product_id", "message": "required",
		})
	}

	if len(r.BillingCycles) == 0 {
		r.Validation.Messages = append(r.Validation.Messages, map[string]interface{}{
			"field": "billing_cycles", "message": "at least one cycle required",
		})
	}

	if r.Status == "" {
		r.Status = enum.PlanStatusActive
	}

	if len(r.Validation.Messages) > 0 {
		return r.Validation.Messages, errors.New(utilities.BadRequest)
	}
	return nil, nil
}
