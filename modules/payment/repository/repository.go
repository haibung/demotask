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
	IPaymentRepository interface {
		Create(ctx context.Context, payment *models.Payment, tx *gorm.DB) (*int, error)
		FindByCaptureID(ctx context.Context, captureID string) (*models.Payment, error)
		// SavePaymentMethod(ctx context.Context, paymentMethod *models.PaymentMethod, tx *gorm.DB) (*int, error)
		// FindPaymentMethodByID(ctx context.Context, id int, userID int) (*models.PaymentMethod, error)
		FindByID(ctx context.Context, id int, userID int) (*models.VaultToken, error)

		FindVaultByUserID(ctx context.Context, userID int) ([]models.VaultToken, error)
	}

	PaymentRepository struct {
		fx.In
		DB     *postgres.DB
		Logger *logger.Logger
	}
)

func NewRepository(repository PaymentRepository) IPaymentRepository {
	return &repository
}

func (r *PaymentRepository) Create(ctx context.Context, payment *models.Payment, tx *gorm.DB) (*int, error) {
	if tx == nil {
		tx = r.DB.Gorm
	}

	var exixting models.Payment
	err := tx.WithContext(ctx).Where("paypal_capture_id = ?", payment.PaypalCaptureID).First(&exixting).Error

	if err == nil {
		return &exixting.ID, nil
	}

	if err != nil && err != gorm.ErrRecordNotFound {
		r.Logger.Error(err)
		return nil, err
	}

	if err := tx.WithContext(ctx).Create(payment).Error; err != nil {
		r.Logger.Error(err)
		return nil, err
	}

	return &payment.ID, nil
}

func (r *PaymentRepository) FindByCaptureID(ctx context.Context, captureID string) (*models.Payment, error) {

	var payment models.Payment

	if err := r.DB.Gorm.WithContext(ctx).Where("paypal_capture_id = ?", captureID).First(&payment).Error; err != nil {
		r.Logger.Error(err)
		return nil, err
	}

	return &payment, nil
}

func (r *PaymentRepository) FindByID(ctx context.Context, id int, userID int) (*models.VaultToken, error) {

	var vaultToken models.VaultToken

	err := r.DB.Gorm.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&vaultToken).Error

	if err != nil {
		return nil, err
	}

	return &vaultToken, nil
}

func (r *PaymentRepository) FindVaultByUserID(ctx context.Context, userID int) ([]models.VaultToken, error) {
	var tokens []models.VaultToken

	err := r.DB.Gorm.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("is_default DESC, id DESC").
		Find(&tokens).Error

	if err != nil {
		return nil, err
	}

	return tokens, nil
}

////
// func (r *PaymentRepository) SavePaymentMethod(ctx context.Context, method *models.PaymentMethod, tx *gorm.DB) (*int, error) {

// 	if tx == nil {
// 		tx = r.DB.Gorm
// 	}

// 	var existing models.PaymentMethod

// 	err := tx.WithContext(ctx).
// 		Where("user_id = ? AND paypal_vault_id = ?", method.UserID, method.PaypalVaultID).
// 		First(&existing).Error

// 	if err == nil {

// 		updates := map[string]interface{}{
// 			"updated_at": time.Now(),
// 		}

// 		if method.Email != "" && existing.Email != method.Email {
// 			updates["email"] = method.Email
// 		}
// 		if method.IsDefault && !existing.IsDefault {
// 			if err := tx.WithContext(ctx).Model(&models.PaymentMethod{}).Where("user_id = ? AND is_default = ?", method.UserID, true).Update("is_default", false).Error; err != nil {
// 				r.Logger.Error(err)
// 				return nil, err
// 			}

// 			updates["is_default"] = true
// 		}

// 		if err := tx.WithContext(ctx).Model(&existing).Updates(updates).Error; err != nil {
// 			r.Logger.Error(err)
// 			return nil, err
// 		}

// 		return &existing.ID, nil
// 	}

// 	if err != nil && err != gorm.ErrRecordNotFound {
// 		r.Logger.Error(err)
// 		return nil, err
// 	}

// 	if method.IsDefault {
// 		if err := tx.WithContext(ctx).Model(&models.PaymentMethod{}).Where("user_id = ? AND is_default = ?", method.UserID, true).Update("is_default", false).Error; err != nil {
// 			r.Logger.Error(err)
// 			return nil, err
// 		}
// 	}

// 	if err := tx.WithContext(ctx).Create(method).Error; err != nil {
// 		r.Logger.Error(err)
// 		return nil, err
// 	}

// 	return &method.ID, nil
// }

// func (r *PaymentRepository) FindPaymentMethodByID(ctx context.Context, id int, userID int) (*models.PaymentMethod, error) {

// 	var method models.PaymentMethod

// 	err := r.DB.Gorm.WithContext(ctx).
// 		Where("id = ? AND user_id = ?", id, userID).
// 		First(&method).Error

// 	if err != nil {
// 		r.Logger.Error(err)
// 		return nil, err
// 	}

// 	return &method, nil
// }

// // func (r *PaymentRepository) FindPaymentMethodByID(ctx context.Context, id int, userID int) (*models.PaymentMethod, error) {

// // 	var method models.PaymentMethod

// // 	err := r.DB.Gorm.WithContext(ctx).
// // 		Where("id = ? AND user_id = ?", id, userID).
// // 		First(&method).Error

// // 	if err != nil {
// // 		r.Logger.Error(err)
// // 		return nil, err
// // 	}

// // 	return &method, nil
// // }
