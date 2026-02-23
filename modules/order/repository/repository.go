package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/demotask/backend/enum"
	"github.com/demotask/backend/models"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/postgres"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type (
	IOrderRepository interface {
		CreateOrder(ctx context.Context, order *models.Order, tx *gorm.DB) (*int, error)
		CreateOrderItems(ctx context.Context, items []models.OrderItem, tx *gorm.DB) error

		// FindByID(ctx context.Context, req *models.Order) (*models.Order, error)
		FindByID(ctx context.Context, id int) (*models.Order, error)
		FindOrderID(ctx context.Context, req *models.Order) (*models.Order, error)
		FindByPaypalOrderID(ctx context.Context, order *models.Order) (*models.Order, error)

		UpdatePaypalOrderID(ctx context.Context, orderID int, paypalOrderID string, tx *gorm.DB) error
		UpdateStatus(ctx context.Context, orderID int, status enum.OrderStatusInternal, tx *gorm.DB) error
		FindOrderPaypalID(ctx context.Context, paypalOrderID string) (*models.Order, error)

		// FindByIDWithPayment
		FindByIDWithPayment(ctx context.Context, orderID int) (*models.Order, *models.Payment, error)
		FindByUserID(ctx context.Context, userID int) ([]models.Order, error)

		// FindAllWithRelationsByUserID
		FindAllWithRelationsByUserID(ctx context.Context, userID int) ([]models.Order, error)
	}

	OrderRepository struct {
		fx.In
		Logger *logger.Logger
		DB     *postgres.DB
	}
)

func NewRepository(repo OrderRepository) IOrderRepository {
	return &repo
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order *models.Order, tx *gorm.DB) (*int, error) {
	if tx == nil {
		tx = r.DB.Gorm
	}
	if err := tx.WithContext(ctx).Create(order).Error; err != nil {
		r.Logger.Error(err)
		return nil, err
	}
	return &order.ID, nil
}

func (r *OrderRepository) CreateOrderItems(ctx context.Context, items []models.OrderItem, tx *gorm.DB) error {
	if len(items) == 0 {
		return nil
	}
	if tx == nil {
		tx = r.DB.Gorm
	}
	if err := tx.WithContext(ctx).Create(items).Error; err != nil {
		r.Logger.Error(err)
		return err
	}
	return nil
}

func (r *OrderRepository) FindOrderID(ctx context.Context, req *models.Order) (*models.Order, error) {
	var order models.Order

	if err := r.DB.Gorm.WithContext(ctx).
		Preload("Items").
		Where("id = ?", req.ID).
		First(&order).Error; err != nil {
		r.Logger.Error(err)
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) FindByID(ctx context.Context, id int) (*models.Order, error) {

	var order models.Order

	if err := r.DB.Gorm.WithContext(ctx).
		Preload("Items").
		Where("id = ?", id).
		First(&order).Error; err != nil {

		r.Logger.Error(err)
		return nil, err
	}

	return &order, nil
}

func (r *OrderRepository) FindByPaypalOrderID(ctx context.Context, req *models.Order) (*models.Order, error) {
	var order models.Order

	if err := r.DB.Gorm.WithContext(ctx).
		Preload("Items").
		Where("paypal_order_id = ?", req.PaypalOrderID).
		First(&order).Error; err != nil {
		r.Logger.Error(err)
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) FindOrderPaypalID(ctx context.Context, paypalOrderID string) (*models.Order, error) {

	if paypalOrderID == "" {
		return nil, fmt.Errorf("paypal_order_id is required")
	}

	var order models.Order

	err := r.DB.Gorm.WithContext(ctx).
		Preload("Items").
		Where("paypal_order_id = ?", paypalOrderID).
		First(&order).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.Logger.Error(err)
		return nil, err
	}

	return &order, nil
}

func (r *OrderRepository) UpdatePaypalOrderID(ctx context.Context, orderID int, paypalOrderID string, tx *gorm.DB) error {
	if tx == nil {
		tx = r.DB.Gorm
	}
	if err := tx.WithContext(ctx).Model(&models.Order{}).
		Where("id = ?", orderID).
		Update("paypal_order_id", paypalOrderID).Error; err != nil {
		r.Logger.Error(err)
		return err
	}
	return nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID int, status enum.OrderStatusInternal, tx *gorm.DB) error {
	if tx == nil {
		tx = r.DB.Gorm
	}
	if err := tx.WithContext(ctx).Model(&models.Order{}).Where("id = ?", orderID).Update("status", status).Error; err != nil {
		r.Logger.Error(err)
		return err
	}
	return nil
}

func (r *OrderRepository) FindByIDWithPayment(ctx context.Context, orderID int) (*models.Order, *models.Payment, error) {

	var order models.Order

	err := r.DB.Gorm.WithContext(ctx).
		Preload("Items").
		First(&order, orderID).Error

	if err != nil {
		return nil, nil, err
	}

	var payment models.Payment
	err = r.DB.Gorm.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("id desc").
		First(&payment).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &order, nil, nil
		}
		return nil, nil, err
	}

	return &order, &payment, nil
}

func (r *OrderRepository) FindByUserID(ctx context.Context, userID int) ([]models.Order, error) {
	var orders []models.Order
	if err := r.DB.Gorm.WithContext(ctx).
		Preload("Items").
		Where("user_id = ?", userID).
		Order("id desc").
		Find(&orders).Error; err != nil {
		r.Logger.Error(err)
		return nil, err
	}
	return orders, nil
}

func (r *OrderRepository) FindAllWithRelationsByUserID(ctx context.Context, userID int) ([]models.Order, error) {

	var orders []models.Order

	err := r.DB.Gorm.WithContext(ctx).
		Preload("Items").
		Preload("Subscription").
		Preload("Subscription.BillingPlan").
		Preload("Subscription.BillingPlan.BillingCycles").
		Where("user_id = ?", userID).
		Order("id desc").
		Find(&orders).Error

	if err != nil {
		r.Logger.Error(err)
		return nil, err
	}

	return orders, nil
}
