package product

import (
	"errors"
	"time"

	"github.com/demotask/backend/enum"
	"github.com/demotask/backend/models"
	"github.com/demotask/backend/packages/validation"
	"github.com/demotask/backend/utilities"
)

type (
	CreateProductRequest struct {
		Name         string            `json:"name"`
		Description  string            `json:"description,omitempty"`
		BusinessType enum.BusinessType `json:"business_type"` // PHYSICAL, DIGITAL, SERVICE

		Category string `json:"category,omitempty"`
		ImageURL string `json:"image_url,omitempty"`
		HomeURL  string `json:"home_url,omitempty"`

		Validation validation.Validation `json:"-"`
	}

	ProductResponse struct {
		ID int `json:"id"`

		Name         string            `json:"name"`
		Description  string            `json:"description,omitempty"`
		BusinessType enum.BusinessType `json:"business_type"`

		Category string `json:"category,omitempty"`
		ImageURL string `json:"image_url,omitempty"`
		HomeURL  string `json:"home_url,omitempty"`

		PaypalProductID *string `json:"paypal_product_id,omitempty"`

		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	CreateOneTimePriceRequest struct {
		ContextUserID *int
		ProductID     int     `json:"product_id"`
		Price         float64 `json:"price"`
		Currency      string  `json:"currency,omitempty"`

		Validation validation.Validation `json:"-"`
	}

	CreateOneTimePriceResponse struct {
		ID        int     `json:"id"`
		ProductID int     `json:"product_id"`
		Price     float64 `json:"price"`
		Currency  string  `json:"currency"`
		IsActive  bool    `json:"is_active"`

		CreatedAt time.Time `json:"created_at"`
	}

	FindProductRequest struct {
		ContextUserID *int

		ID int
	}

	UpdateProductRequest struct {
		Name        *string `json:"name,omitempty"`
		Description *string `json:"description,omitempty"`
		Category    *string `json:"category,omitempty"`
		ImageURL    *string `json:"image_url,omitempty"`
		HomeURL     *string `json:"home_url,omitempty"`

		Validation validation.Validation `json:"-"`
	}

	ProductPricingResponse struct {
		Product             ProductResponse              `json:"product"`
		PricingTypes        []string                     `json:"pricing_types,omitempty"`
		OneTimePrice        *OneTimePriceResponse        `json:"one_time_price,omitempty"`
		SubscriptionOptions []SubscriptionOptionResponse `json:"subscription_options,omitempty"`
	}

	OneTimePriceResponse struct {
		Price    float64 `json:"price"`
		Currency string  `json:"currency"`
	}

	SubscriptionOptionResponse struct {
		BillingPlanID int     `json:"billing_plan_id"`
		IntervalUnit  string  `json:"interval_unit"`
		IntervalCount int     `json:"interval_count"`
		BasePrice     float64 `json:"base_price"`
		FinalPrice    float64 `json:"final_price"`
		Currency      string  `json:"currency"`

		HasTrial           bool   `json:"has_trial"`
		TrialIntervalUnit  string `json:"trial_interval_unit,omitempty"`
		TrialIntervalCount int    `json:"trial_interval_count,omitempty"`
		TrialCycles        int    `json:"trial_cycles,omitempty"`

		TaxPercentage *float64 `json:"tax_percentage,omitempty"`
		TaxInclusive  *bool    `json:"tax_inclusive,omitempty"`
	}

	ListProductsResponse struct {
		Items []ProductResponse `json:"items"`
	}
)

func (r *CreateProductRequest) Validate() ([]map[string]interface{}, error) {
	r.Validation.IsEmptyString(r.Name, "name")
	r.Validation.IsEmptyString(string(r.BusinessType), "business_type")

	if err := r.BusinessType.IsValid(); err != nil {
		r.Validation.Messages = append(r.Validation.Messages, map[string]interface{}{
			"field":   "business_type",
			"message": "must be PHYSICAL, DIGITAL, or SERVICE",
		})
	}

	if len(r.Validation.Messages) > 0 {
		return r.Validation.Messages, errors.New(utilities.BadRequest)
	}
	return nil, nil
}

func (r *CreateOneTimePriceRequest) Validate() ([]map[string]interface{}, error) {
	r.Validation.IsEmptyString(r.Currency, "currency")
	r.Validation.IsFloatMin(r.Price, 0, "Price")

	if len(r.Validation.Messages) > 0 {
		return r.Validation.Messages, errors.New(utilities.BadRequest)
	}
	return nil, nil
}

// func (receiver *CreatePlanRequest) Validate() ([]map[string]interface{}, error) {
// 	receiver.Validation.IsEmptyString(receiver.Name, "Name")

// 	if receiver.ProductID <= 0 {
// 		receiver.Validation.Messages = append(receiver.Validation.Messages, map[string]interface{}{
// 			"field":   "ProductID",
// 			"message": "Product ID is required",
// 		})
// 	}

// 	if len(receiver.BillingCycles) == 0 {
// 		receiver.Validation.Messages = append(receiver.Validation.Messages, map[string]interface{}{
// 			"field":   "BillingCycles",
// 			"message": "At least one billing cycle is required",
// 		})
// 	}

// 	// Set default status
// 	if receiver.Status == "" {
// 		receiver.Status = "ACTIVE"
// 	}

// 	if len(receiver.Validation.Messages) > 0 {
// 		return receiver.Validation.Messages, errors.New(utilities.BadRequest)
// 	}

// 	return nil, nil
// }

type GetProductRequest *models.Product

type FindAllRequest struct {
	UserID *int
}
