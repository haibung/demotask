package models

import (
	"time"

	"github.com/demotask/backend/enum"
	"gorm.io/gorm"
)

type Order struct {
	ID            int     `gorm:"primaryKey;column:id"`
	UserID        int     `gorm:"column:user_id;not null"`
	PaypalOrderID *string `gorm:"column:paypal_order_id"`

	Status    enum.OrderStatusInternal `gorm:"column:status"`
	OrderType enum.OrderType           `gorm:"column:order_type"`
	Snapshot  string

	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Read-only snapshot
	Items        []OrderItem   `gorm:"<-:false;foreignKey:OrderID;references:ID"`
	Subscription *Subscription `gorm:"<-:false;foreignKey:ID;references:OrderID"`

	// Attribute
	OrderSnapshot OrderSnapshot `gorm:"-"`
}

type OrderSnapshot struct {
	Items    []OrderItemSnapshot
	Subtotal float64
	Total    float64
	Currency string

	PaymentMethod string
	Intent        string
}

type OrderItemSnapshot struct {
	ProductID    int
	ProductName  string
	Quantity     int
	UnitPrice    float64
	TotalPrice   float64
	Currency     string
	BusinessType string // ONE TIME | SUBSCRIPTION
	PlanID       *string
}
