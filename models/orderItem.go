package models

import (
	"time"

	"gorm.io/gorm"
)

type OrderItem struct {
	ID        int `gorm:"primaryKey;column:id"`
	OrderID   int `gorm:"column:order_id;not null"`
	ProductID int `gorm:"column:product_id;not null"`

	ProductName string  `gorm:"column:product_name;not null"`
	Quantity    int     `gorm:"column:quantity;not null"`
	UnitPrice   float64 `gorm:"column:unit_price;not null"`
	TotalPrice  float64 `gorm:"column:total_price;not null"`
	Currency    string  `gorm:"column:currency;not null"`

	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Order *Order `gorm:"<-:false;foreignKey:OrderID;references:ID"`
}
