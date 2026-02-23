package models

import (
	"time"

	"gorm.io/gorm"
)

type BillingCycle struct {
	ID            int `gorm:"primaryKey;column:id"`
	BillingPlanID int `gorm:"column:billing_plan_id;not null"`

	IntervalUnit  string `gorm:"column:interval_unit;not null"`  // DAY, WEEK, MONTH, YEAR
	IntervalCount int    `gorm:"column:interval_count;not null"` // 1, 3, 6, 12

	TenureType  string `gorm:"column:tenure_type;not null"` // TRIAL, REGULAR
	Sequence    int    `gorm:"column:sequence;not null"`
	TotalCycles int    `gorm:"column:total_cycles;not null"` // 0 = unlimited

	PriceValue float64 `gorm:"column:price_value;not null"`
	Currency   string  `gorm:"column:currency;not null"`

	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Read-only relation
	BillingPlan *BillingPlan `gorm:"<-:false;foreignKey:BillingPlanID;references:ID"`
}
