package models

import (
	"time"

	"github.com/demotask/backend/enum"
	"gorm.io/gorm"
)

type Payment struct {
	ID                   int                `gorm:"column:id"`
	UserID               int                `gorm:"column:user_id"`
	OrderID              *int               `gorm:"column:order_id"`
	SubscriptionID       *int               `gorm:"column:subscription_id"`
	Provider             string             `gorm:"column:provider"`
	PaypalOrderID        *string            `gorm:"column:paypal_order_id"`
	PaypalCaptureID      *string            `gorm:"column:paypal_capture_id"`
	PaypalSubscriptionID *string            `gorm:"column:paypal_subscription_id"`
	Amount               float64            `gorm:"column:amount"`
	Currency             string             `gorm:"column:currency"`
	Status               enum.PaymentStatus `gorm:"column:status"`
	RawResponse          string             `gorm:"column:raw_response"`

	PaidAt *time.Time `gorm:"column:paid_at"`

	Order Order `gorm:"<-:false;foreignKey:OrderID;references:ID"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}
