package repository

import (
	"context"

	"github.com/demotask/backend/models"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/postgres"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type (
	IUserInterface interface {
		Create(ctx context.Context, reqData *models.User, tx *gorm.DB) (*int, error)
		FindByID(ctx context.Context, reqData *models.User) (*models.User, error)
		FindByEmail(ctx context.Context, reqData *models.User) (*models.User, error)
	}

	UserRepository struct {
		fx.In
		DB     *postgres.DB
		Logger *logger.Logger
	}
)

// NewRepository :
func NewRepository(userRepository UserRepository) IUserInterface {
	return &userRepository
}

// Create :
func (receiver *UserRepository) Create(ctx context.Context, reqData *models.User, tx *gorm.DB) (*int, error) {
	if tx == nil {
		tx = receiver.DB.Gorm
	}
	if err := tx.WithContext(ctx).Create(&reqData).Error; err != nil {
		receiver.Logger.Error(err)
		return nil, err
	}

	return &reqData.ID, nil
}

// FindByID :
func (receiver *UserRepository) FindByID(ctx context.Context, reqData *models.User) (*models.User, error) {
	user := new(models.User)

	if err := receiver.DB.Gorm.WithContext(ctx).
		// Preload("FetchAllPermissionByRoleID.FetchFeatureByFeatureID").
		Where(&models.User{
			ID: reqData.ID,
		}).
		First(&user).Error; err != nil {
		receiver.Logger.Error(err)
		return nil, err
	}

	return user, nil
}

// FindByEmail :
func (receiver *UserRepository) FindByEmail(ctx context.Context, reqData *models.User) (*models.User, error) {
	user := new(models.User)

	if err := receiver.DB.Gorm.WithContext(ctx).
		Where(&models.User{
			Email: reqData.Email,
		}).
		First(&user).Error; err != nil {
		receiver.Logger.Error(err)
		return nil, err
	}

	return user, nil
}
