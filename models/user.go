package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID       int    `gorm:"primaryKey;column:id"`
	FullName string `gorm:"column:full_name;not null"`
	Password string `gorm:"column:password;not null"`
	Email    string `gorm:"column:email;unique;not null"`

	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Read-only relations
	Orders        []Order        `gorm:"<-:false;foreignKey:UserID;references:ID"`
	Subscriptions []Subscription `gorm:"<-:false;foreignKey:UserID;references:ID"`
	VaultTokens   []VaultToken   `gorm:"<-:false;foreignKey:UserID;references:ID"`
}
