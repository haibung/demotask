package payment

import (
	"errors"

	"github.com/demotask/backend/packages/validation"
	"github.com/demotask/backend/utilities"
)

type (
	CapturePaymentRequest struct {
		ContextUserID *int   `json:"-"`
		PaypalOrderID string `json:"paypal_order_id"`

		Validation validation.Validation
	}

	CapturePaymentResponse struct {
		ID              int    `json:"id"`
		OrderID         int    `json:"order_id"`
		PaypalOrderID   string `json:"paypal_order_id"`
		PaypalCaptureID string `json:"paypal_capture_id"`

		Status string `json:"status"`
	}

	PayAgainRequest struct {
		ContextUserID *int `json:"-"`
		OrderID       int  `json:"order_id"`
		VaultTokenID  int  `json:"vault_token_id"`

		Validation validation.Validation
	}

	PayAgainResponse struct {
		PaymentID       int    `json:"payment_id"`
		OrderID         int    `json:"order_id"`
		PaypalOrderID   string `json:"paypal_order_id"`
		PaypalCaptureID string `json:"paypal_capture_id"`
		Status          string `json:"status"`
	}

	GetVaultTokensRequest struct {
		ContextUserID *int `json:"-"`
	}

	VaultTokenItem struct {
		ID        int     `json:"id"`
		Email     *string `json:"email,omitempty"`
		IsDefault bool    `json:"is_default"`
	}

	GetVaultTokensResponse struct {
		Items []VaultTokenItem `json:"items"`
	}
)

func (r *CapturePaymentRequest) Validate() ([]map[string]interface{}, error) {

	if r.ContextUserID == nil {
		r.Validation.Messages = append(r.Validation.Messages, map[string]interface{}{
			"user": "unauthorized",
		})
	}

	r.Validation.IsEmptyString(r.PaypalOrderID, "paypal_order_id")

	if len(r.Validation.Messages) > 0 {
		return r.Validation.Messages, errors.New(utilities.BadRequest)
	}

	return nil, nil
}

func (r *PayAgainRequest) Validate() ([]map[string]interface{}, error) {

	if r.ContextUserID == nil {
		r.Validation.Messages = append(r.Validation.Messages, map[string]interface{}{
			"user": "unauthorized",
		})
	}

	if r.OrderID <= 0 {
		r.Validation.Messages = append(r.Validation.Messages, map[string]interface{}{
			"order_id": "order_id is required",
		})
	}

	if r.VaultTokenID <= 0 {
		r.Validation.Messages = append(r.Validation.Messages, map[string]interface{}{
			"vault_token_id": "vault_token_id is required",
		})
	}

	if len(r.Validation.Messages) > 0 {
		return r.Validation.Messages, errors.New(utilities.BadRequest)
	}

	return nil, nil
}
