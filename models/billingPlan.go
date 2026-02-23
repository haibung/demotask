package models

import (
	"time"

	"github.com/demotask/backend/enum"
	"gorm.io/gorm"
)

type BillingPlan struct {
	ID           int     `gorm:"primaryKey;column:id"`
	ProductID    int     `gorm:"column:product_id;not null"`
	PaypalPlanID *string `gorm:"column:paypal_plan_id"`

	Name        string  `gorm:"column:name;not null"`
	Description *string `gorm:"column:description"`
	Status      enum.PlanStatus

	AutoBillOutstanding     bool `gorm:"column:auto_bill_outstanding;default:true"`
	PaymentFailureThreshold int  `gorm:"column:payment_failure_threshold;default:0"`

	TaxPercentage *float64 `gorm:"column:tax_percentage"`
	TaxInclusive  *bool    `gorm:"column:tax_inclusive"`

	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Read-only relations
	Product       *Product       `gorm:"<-:false;foreignKey:ProductID;references:ID"`
	BillingCycles []BillingCycle `gorm:"<-:false;foreignKey:BillingPlanID;references:ID"`
}
