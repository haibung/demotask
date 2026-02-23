package models

import (
	"time"

	"gorm.io/gorm"
)

type VaultToken struct {
	ID            int     `gorm:"primaryKey;column:id"`
	UserID        int     `gorm:"column:user_id;not null"`
	PayPalVaultID string  `gorm:"column:paypal_vault_id;unique;not null"`
	Email         *string `gorm:"column:email"`
	IsDefault     bool    `gorm:"column:is_default;default:false"`

	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Read-only relation
	User *User `gorm:"<-:false;foreignKey:UserID;references:ID"`
}
