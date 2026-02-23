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
	IProductRepository interface {
		Create(ctx context.Context, reqData *models.Product, tx *gorm.DB) (*int, error)
		Find(ctx context.Context, reqData *models.Product) (*models.Product, error)
		CreateOneTimePrice(ctx context.Context, reqData *models.ProductOneTimePrice, tx *gorm.DB) (*int, error)
		FindAll(ctx context.Context, filter *models.Product) ([]models.Product, error)
		FindByIDs(ctx context.Context, ids []int) ([]models.Product, error)
		UpdatePlanID(ctx context.Context, productID int64, planID string, tx *gorm.DB) error
		FindAllWithPrice(ctx context.Context) ([]models.Product, error)
	}

	ProductRepository struct {
		fx.In
		Logger *logger.Logger
		DB     *postgres.DB
	}
)

func NewRepository(productRepository ProductRepository) IProductRepository {
	return &productRepository
}

func (l *ProductRepository) Create(ctx context.Context, reqData *models.Product, tx *gorm.DB) (*int, error) {
	if tx == nil {
		tx = l.DB.Gorm
	}
	if err := tx.WithContext(ctx).Create(reqData).Error; err != nil {
		l.Logger.Error(err)
		return nil, err
	}
	return &reqData.ID, nil
}

func (l *ProductRepository) Find(ctx context.Context, reqData *models.Product) (*models.Product, error) {
	var product models.Product
	err := l.DB.Gorm.WithContext(ctx).
		Where(&models.Product{
			ID: reqData.ID,
		}).
		First(&product).Error
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}
	return &product, err
}

func (l *ProductRepository) CreateOneTimePrice(ctx context.Context, reqData *models.ProductOneTimePrice, tx *gorm.DB) (*int, error) {
	if tx == nil {
		tx = l.DB.Gorm
	}
	if err := tx.WithContext(ctx).Create(reqData).Error; err != nil {
		l.Logger.Error(err)
		return nil, err
	}
	return &reqData.ID, nil
}

func (l *ProductRepository) FindAll(ctx context.Context, filter *models.Product) ([]models.Product, error) {

	var products []models.Product

	query := l.DB.Gorm.WithContext(ctx).
		Preload("OneTimePrices", "is_active = ?", true).
		Preload("BillingPlans", "status = ?", enum.PlanStatusActive).
		Preload("BillingPlans.BillingCycles")

	if filter != nil {
		query = query.Where(filter)
	}

	if err := query.Find(&products).Error; err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	return products, nil
}

func (l *ProductRepository) FindByIDs(ctx context.Context, ids []int) ([]models.Product, error) {
	var products []models.Product
	if err := l.DB.Gorm.WithContext(ctx).
		Preload("OneTimePrices", "deleted_at IS NULL").
		Where("id IN ?", ids).
		Find(&products).Error; err != nil {
		l.Logger.Error(err)
		return nil, err
	}
	return products, nil
}

func (l *ProductRepository) UpdatePlanID(ctx context.Context, productID int64, planID string, tx *gorm.DB) error {
	if tx == nil {
		tx = l.DB.Gorm
	}
	if err := tx.WithContext(ctx).Model(&models.Product{}).Where("id = ?", productID).Update("paypal_plan_id", planID).Error; err != nil {
		l.Logger.Error(err)
		return err
	}
	return nil
}

func (l *ProductRepository) FindAllWithPrice(ctx context.Context) ([]models.Product, error) {
	var products []models.Product

	if err := l.DB.Gorm.WithContext(ctx).
		Preload("OneTimePrices", "is_active = ?", true).
		Where("deleted_at IS NULL").
		Find(&products).Error; err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	return products, nil
}
