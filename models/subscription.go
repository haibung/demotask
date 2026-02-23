package models

import (
	"time"

	"github.com/demotask/backend/enum"
	"gorm.io/gorm"
)

type Subscription struct {
	ID                   int
	UserID               int
	OrderID              int
	BillingPlanID        int
	PaypalSubscriptionID string
	Status               enum.SubscriptionStatus
	StartTime            *time.Time
	Snapshot             string
	Quantity             int

	NextBillingTime     *time.Time
	FailedPaymentsCount int
	LastPaymentTime     *time.Time

	VaultTokenID *int

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Read-only
	BillingPlan *BillingPlan `gorm:"<-:false;foreignKey:BillingPlanID;references:ID"`
	Order       *Order       `gorm:"<-:false;foreignKey:OrderID;references:ID"`

	// Virtual attribute
	SubscriptionSnapshot SubscriptionSnapshot `gorm:"<-:false;-;"`
}

type SubscriptionSnapshot struct {
	Type          string
	BillingPlanID int
	PlanName      string
	PaypalPlanID  string
	IntervalUnit  string
	IntervalCount int
	Price         float64
	Currency      string
	Quantity      int
	TotalPrice    float64
	// StartTime     *time.Time
}
