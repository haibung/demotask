package models

import (
	"time"

	"github.com/demotask/backend/enum"
	"gorm.io/gorm"
)

type WebhookEvent struct {
	ID          int
	Provider    string
	EventType   string
	EventID     string `gorm:"uniqueIndex"`
	Payload     string `gorm:"type:jsonb"`
	ProcessedAt *time.Time
	Status      enum.WebhookStatus

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}
