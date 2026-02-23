package repository

import (
	"context"
	"time"

	"github.com/demotask/backend/enum"
	"github.com/demotask/backend/models"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/postgres"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type IWebhookRepository interface {
	FindByEventID(ctx context.Context, eventID string) (*models.WebhookEvent, error)
	Save(ctx context.Context, event *models.WebhookEvent, tx *gorm.DB) (*int, error)
	MarkProcessed(ctx context.Context, id int, tx *gorm.DB) error
	MarkFailed(ctx context.Context, id int, tx *gorm.DB) error
}

type WebhookRepository struct {
	fx.In
	DB     *postgres.DB
	Logger *logger.Logger
}

func NewRepository(r WebhookRepository) IWebhookRepository {
	return &r
}

func (r *WebhookRepository) FindByEventID(ctx context.Context, eventID string) (*models.WebhookEvent, error) {
	var event models.WebhookEvent
	err := r.DB.Gorm.WithContext(ctx).
		Where("event_id = ?", eventID).
		First(&event).Error

	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *WebhookRepository) Save(ctx context.Context, event *models.WebhookEvent, tx *gorm.DB) (*int, error) {
	if err := tx.WithContext(ctx).Create(event).Error; err != nil {
		return nil, err
	}
	return &event.ID, nil
}

func (r *WebhookRepository) MarkProcessed(ctx context.Context, id int, tx *gorm.DB) error {
	now := time.Now()
	return tx.WithContext(ctx).
		Model(&models.WebhookEvent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       enum.WebhookStatusProcessed.String(),
			"processed_at": &now,
		}).Error
}

func (r *WebhookRepository) MarkFailed(ctx context.Context, id int, tx *gorm.DB) error {
	return tx.WithContext(ctx).
		Model(&models.WebhookEvent{}).
		Where("id = ?", id).
		Update("status", enum.WebhookStatusFailed.String()).Error
}
