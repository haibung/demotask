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
	IBillingPlanRepository interface {
		Create(ctx context.Context, req *models.BillingPlan, tx *gorm.DB) (*int, error)
		CreateBillingCycle(ctx context.Context, req *models.BillingCycle, tx *gorm.DB) error
		FindByID(ctx context.Context, req *models.BillingPlan) (*models.BillingPlan, error)
		FindByPayPalID(ctx context.Context, paypalSubscriptionID string) (*models.Subscription, error)
		UpdateStatus(ctx context.Context, paypalSubscriptionID string, status string, tx *gorm.DB) error
		ExistsActivePlan(ctx context.Context, productID int, name string) (bool, error)
	}

	BillingPlanRepository struct {
		fx.In
		Logger *logger.Logger
		DB     *postgres.DB
	}
)

func NewRepository(repo BillingPlanRepository) IBillingPlanRepository {
	return &repo
}

func (r *BillingPlanRepository) Create(ctx context.Context, req *models.BillingPlan, tx *gorm.DB) (*int, error) {
	if tx == nil {
		tx = r.DB.Gorm
	}
	if err := tx.WithContext(ctx).Create(req).Error; err != nil {
		r.Logger.Error(err)
		return nil, err
	}
	return &req.ID, nil
}

func (r *BillingPlanRepository) CreateBillingCycle(ctx context.Context, req *models.BillingCycle, tx *gorm.DB) error {
	if tx == nil {
		tx = r.DB.Gorm
	}
	return tx.WithContext(ctx).Create(req).Error
}

func (r *BillingPlanRepository) FindByID(ctx context.Context, req *models.BillingPlan) (*models.BillingPlan, error) {
	var plan models.BillingPlan
	if err := r.DB.Gorm.WithContext(ctx).Preload("BillingCycles").
		Where("id = ?", req.ID).
		First(&plan).Error; err != nil {
		r.Logger.Error(err)
		return nil, err
	}
	return &plan, nil
}

func (r *BillingPlanRepository) FindByPayPalID(ctx context.Context, paypalSubscriptionID string) (*models.Subscription, error) {
	var subscription models.Subscription
	if err := r.DB.Gorm.WithContext(ctx).Where("paypal_subscription_id = ?", paypalSubscriptionID).First(&subscription).Error; err != nil {
		r.Logger.Error(err)
		return nil, err
	}
	return &subscription, nil
}

func (r *BillingPlanRepository) UpdateStatus(ctx context.Context, paypalSubscriptionID string, status string, tx *gorm.DB) error {
	if tx == nil {
		tx = r.DB.Gorm
	}
	if err := tx.WithContext(ctx).Model(&models.Subscription{}).Where("paypal_subscription_id = ?", paypalSubscriptionID).Update("status", status).Error; err != nil {
		r.Logger.Error(err)
		return err
	}
	return nil
}

func (r *BillingPlanRepository) ExistsActivePlan(ctx context.Context, productID int, name string) (bool, error) {
	var count int64
	if err := r.DB.Gorm.WithContext(ctx).Model(&models.BillingPlan{}).Where("product_id = ? AND name = ? AND status = ?", productID, name, enum.PlanStatusActive).Count(&count).Error; err != nil {
		r.Logger.Error(err)
		return false, err
	}
	return count > 0, nil
}
