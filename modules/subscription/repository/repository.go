package repository

import (
	"context"

	"github.com/demotask/backend/enum"
	"github.com/demotask/backend/models"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/postgres"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type (
	ISubscriptionRepository interface {
		Create(ctx context.Context, subscription *models.Subscription, tx *gorm.DB) (*int, error)
		FindByPayPalID(ctx context.Context, paypalSubscriptionID string) (*models.Subscription, error)
		UpdateStatus(ctx context.Context, paypalSubscriptionID string, status string, tx *gorm.DB) error
		UpdateStartTime(ctx context.Context, paypalSubscriptionID string, tx *gorm.DB) error
		FindByID(ctx context.Context, id int) (*models.Subscription, error)
		FindByUserID(ctx context.Context, userID int) ([]models.Subscription, error)
		FindByOrderID(ctx context.Context, orderID int) (*models.Subscription, error)
	}

	SubscriptionRepository struct {
		fx.In
		Logger *logger.Logger
		DB     *postgres.DB
	}
)

func NewRepository(repo SubscriptionRepository) ISubscriptionRepository {
	return &repo
}

func (r *SubscriptionRepository) Create(ctx context.Context, subscription *models.Subscription, tx *gorm.DB) (*int, error) {
	if tx == nil {
		tx = r.DB.Gorm
	}
	if err := tx.WithContext(ctx).Create(subscription).Error; err != nil {
		r.Logger.Error(err)
		return nil, err
	}
	return &subscription.ID, nil
}

func (r *SubscriptionRepository) FindByPayPalID(ctx context.Context, paypalSubscriptionID string) (*models.Subscription, error) {
	var subscription models.Subscription
	if err := r.DB.Gorm.WithContext(ctx).Where("paypal_subscription_id = ?", paypalSubscriptionID).First(&subscription).Error; err != nil {
		r.Logger.Error(err)
		return nil, err
	}
	return &subscription, nil
}

func (r *SubscriptionRepository) UpdateStatus(ctx context.Context, paypalSubscriptionID string, status string, tx *gorm.DB) error {
	if tx == nil {
		tx = r.DB.Gorm
	}
	if err := tx.WithContext(ctx).Model(&models.Subscription{}).Where("paypal_subscription_id = ?", paypalSubscriptionID).Update("status", status).Error; err != nil {
		r.Logger.Error(err)
		return err
	}
	return nil
}

func (r *SubscriptionRepository) FindByID(ctx context.Context, id int) (*models.Subscription, error) {

	var sub models.Subscription

	if err := r.DB.Gorm.WithContext(ctx).Preload("BillingPlan").First(&sub, id).Error; err != nil {
		r.Logger.Error(err)
		return nil, err
	}

	return &sub, nil
}

func (r *SubscriptionRepository) FindByUserID(ctx context.Context, userID int) ([]models.Subscription, error) {

	var subs []models.Subscription

	if err := r.DB.Gorm.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, enum.SubscriptionStatusActive).
		Preload("BillingPlan").
		Find(&subs).Error; err != nil {

		r.Logger.Error(err)
		return nil, err
	}

	return subs, nil
}

func (r *SubscriptionRepository) FindByOrderID(ctx context.Context, orderID int) (*models.Subscription, error) {

	var sub models.Subscription

	if err := r.DB.Gorm.WithContext(ctx).Where("order_id = ?", orderID).First(&sub).Error; err != nil {

		r.Logger.Error(err)
		return nil, err
	}

	return &sub, nil
}

func (r *SubscriptionRepository) UpdateStartTime(ctx context.Context, paypalSubscriptionID string, tx *gorm.DB) error {

	if tx == nil {
		tx = r.DB.Gorm
	}

	if err := tx.WithContext(ctx).Model(&models.Subscription{}).Where("paypal_subscription_id = ?", paypalSubscriptionID).Update("start_time", gorm.Expr("now()")).Error; err != nil {

		r.Logger.Error(err)
		return err
	}

	return nil
}
