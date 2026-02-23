package models

import (
	"time"

	"github.com/demotask/backend/enum"
	"gorm.io/gorm"
)

type Product struct {
	ID              int               `gorm:"primaryKey" json:"id"`
	PaypalProductID string            `json:"paypal_product_id"`
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	BusinessType    enum.BusinessType `gorm:"column:business_type" json:"business_type"`
	Category        string            `json:"category"`
	ImageURL        string            `json:"image_url"`
	HomeURL         string            `json:"home_url"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	DeletedAt       gorm.DeletedAt    `gorm:"index" json:"-"`

	// Relationships
	OneTimePrices []ProductOneTimePrice `gorm:"<-:false;foreignKey:ProductID;references:ID"`
	BillingPlans  []BillingPlan         `gorm:"<-:false;foreignKey:ProductID;references:ID"`
}
