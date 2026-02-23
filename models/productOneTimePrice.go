package models

import (
	"time"

	"gorm.io/gorm"
)

type ProductOneTimePrice struct {
	ID        int     `gorm:"primaryKey;column:id"`
	ProductID int     `gorm:"column:product_id;not null"`
	Price     float64 `gorm:"column:price;not null"`
	Currency  string  `gorm:"column:currency;not null"`
	IsActive  bool    `gorm:"column:is_active;default:true"`

	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Relations
	Product *Product `gorm:"<-:false;foreignKey:ProductID;references:ID"`
}
