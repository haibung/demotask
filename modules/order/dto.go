package order

import (
	"errors"
	"fmt"
	"time"

	"github.com/demotask/backend/packages/validation"
	"github.com/demotask/backend/utilities"
)

type (
	CreateOrderRequest struct {
		ContextUserID *int                     `json:"-"`
		Items         []CreateOrderItemRequest `json:"items"`
		Validation    validation.Validation    `json:"-"`
	}

	CreateOrderItemRequest struct {
		ProductID int `json:"product_id"`
		Quantity  int `json:"quantity"`
	}

	CreateOrderResponse struct {
		OrderID     string
		PaypalID    string `json:"-"`
		Status      string `json:"status"`
		ApprovalURL string `json:"-"`
	}

	OrderSnapshot struct {
		Items      []OrderItemSnapshot
		TotalValue float64
		Currency   string
	}

	OrderItemSnapshot struct {
		ProductID   int
		ProductName string
		Quantity    int
		UnitPrice   float64
		TotalPrice  float64
		Currency    string
	}

	PayOrderRequest struct {
		ContextUserID *int `json:"-"`
		OrderID       int  `json:"order_id"`

		Validation validation.Validation `json:"-"`
	}

	PayOrderResponse struct {
		OrderID       int
		PaypalOrderID string
		Status        string
		ApprovalURL   string
	}

	PayAgainRequest struct {
		ContextUserID   *int
		OrderID         int
		PaymentMethodID int
		Validation      validation.Validation `json:"-"`
	}

	GetOrderResponse struct {
		OrderID    int                 `json:"order_id"`
		Status     string              `json:"status"`
		TotalValue float64             `json:"total_value"`
		Currency   string              `json:"currency"`
		Items      []OrderItemResponse `json:"items"`
		Payment    *PaymentInfo        `json:"payment,omitempty"`
		CreatedAt  time.Time           `json:"created_at"`
	}

	OrderItemResponse struct {
		ProductID   int     `json:"product_id"`
		ProductName string  `json:"product_name"`
		Quantity    int     `json:"quantity"`
		UnitPrice   float64 `json:"unit_price"`
		TotalPrice  float64 `json:"total_price"`
		Currency    string  `json:"currency"`
	}

	PaymentInfo struct {
		Provider        string     `json:"provider"`
		PaypalOrderID   string     `json:"paypal_order_id"`
		PaypalCaptureID string     `json:"paypal_capture_id"`
		Status          string     `json:"status"`
		PaidAt          *time.Time `json:"paid_at,omitempty"`
	}

	GetMyOrdersResponse struct {
		OrderID       int                 `json:"order_id"`
		PaypalOrderID string              `json:"paypal_order_id"`
		OrderType     string              `json:"order_type"`
		Status        string              `json:"status"`
		TotalValue    float64             `json:"total_value"`
		Currency      string              `json:"currency"`
		Items         []OrderItemDetail   `json:"items,omitempty"`
		Subscription  *SubscriptionDetail `json:"subscription,omitempty"`
		CreatedAt     time.Time           `json:"created_at"`
	}

	OrderItemDetail struct {
		ProductID   int     `json:"product_id"`
		ProductName string  `json:"product_name"`
		Quantity    int     `json:"quantity"`
		UnitPrice   float64 `json:"unit_price"`
		TotalPrice  float64 `json:"total_price"`
		Currency    string  `json:"currency"`
	}

	SubscriptionDetail struct {
		SubscriptionID       int        `json:"subscription_id"`
		PaypalSubscriptionID string     `json:"paypal_subscription_id"`
		Status               string     `json:"status"`
		PlanName             string     `json:"plan_name"`
		IntervalUnit         string     `json:"interval_unit"`
		IntervalCount        int        `json:"interval_count"`
		PriceValue           float64    `json:"price_value"`
		Currency             string     `json:"currency"`
		StartTime            *time.Time `json:"start_time,omitempty"`
		NextBillingTime      *time.Time `json:"next_billing_time,omitempty"`
		FailedPaymentsCount  int        `json:"failed_payments_count,omitempty"`
		LastPaymentTime      *time.Time `json:"last_payment_time,omitempty"`
	}
)

func (r *CreateOrderRequest) Validate() ([]map[string]interface{}, error) {
	if len(r.Items) == 0 {
		r.Validation.Messages = append(r.Validation.Messages, map[string]interface{}{
			"items": "at least one item is required",
		})
	}

	for i, item := range r.Items {
		if item.ProductID <= 0 {
			r.Validation.Messages = append(r.Validation.Messages, map[string]interface{}{
				fmt.Sprintf("items[%d].product_id", i): "invalid product id",
			})
		}

		if item.Quantity <= 0 {
			r.Validation.Messages = append(r.Validation.Messages, map[string]interface{}{
				fmt.Sprintf("items[%d].quantity", i): "quantity must be greater than zero",
			})
		}
	}

	if len(r.Validation.Messages) > 0 {
		return r.Validation.Messages, errors.New(utilities.BadRequest)
	}

	return nil, nil
}

func (r *PayOrderRequest) Validate() ([]map[string]interface{}, error) {
	if r.ContextUserID == nil {
		r.Validation.Messages = append(r.Validation.Messages, map[string]interface{}{
			"user": "unauthorized",
		})
	}

	if r.OrderID <= 0 {
		r.Validation.Messages = append(r.Validation.Messages, map[string]interface{}{
			"order_id": "invalid order id",
		})
	}

	if len(r.Validation.Messages) > 0 {
		return r.Validation.Messages, errors.New(utilities.BadRequest)
	}

	return nil, nil
}

func (r *PayAgainRequest) Validate() ([]map[string]interface{}, error) {

	if r.ContextUserID == nil {
		r.Validation.Messages = append(
			r.Validation.Messages,
			map[string]interface{}{"user": "unauthorized"},
		)
	}

	r.Validation.IsIntegerMin(r.OrderID, 1, "order_id")
	r.Validation.IsIntegerMin(r.PaymentMethodID, 1, "payment_method_id")

	if len(r.Validation.Messages) > 0 {
		return r.Validation.Messages, errors.New(utilities.BadRequest)
	}

	return nil, nil
}
