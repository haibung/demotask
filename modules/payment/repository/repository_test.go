package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/demotask/backend/enum"
	"github.com/demotask/backend/models"
	"github.com/demotask/backend/modules/payment/repository"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/postgres"
	"github.com/stretchr/testify/assert"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*postgres.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(gormPostgres.New(gormPostgres.Config{
		Conn: db,
	}), &gorm.Config{})
	assert.NoError(t, err)

	postgresDB := &postgres.DB{
		Sql:  db,
		Gorm: gormDB,
	}

	return postgresDB, mock
}

func setupRepository(t *testing.T) (repository.IPaymentRepository, sqlmock.Sqlmock) {
	db, mock := setupTestDB(t)
	log := logger.NewLogger()

	repo := repository.NewRepository(repository.PaymentRepository{
		DB:     db,
		Logger: log,
	})

	return repo, mock
}

func TestCreatePayment(t *testing.T) {
	repo, mock := setupRepository(t)

	now := time.Now()
	orderID := 1
	captureID := "CAP-123"
	payment := &models.Payment{
		OrderID:         &orderID,
		Provider:        "PAYPAL",
		PaypalCaptureID: &captureID,
		Amount:          100.0,
		Currency:        "USD",
		Status:          enum.PaymentStatusCaptured,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	mock.ExpectQuery(`^SELECT \* FROM "payments" WHERE paypal_capture_id = \$1 AND "payments"\."deleted_at" IS NULL ORDER BY "payments"\."id" LIMIT \$2`).
		WithArgs("CAP-123", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectQuery(`^INSERT INTO "payments"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	id, err := repo.Create(context.Background(), payment, nil)
	assert.NoError(t, err)
	if id != nil {
		assert.Equal(t, 1, *id)
	}

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestFindByCaptureID(t *testing.T) {
	repo, mock := setupRepository(t)

	mock.ExpectQuery(`^SELECT \* FROM "payments" WHERE paypal_capture_id = \$1 AND "payments"\."deleted_at" IS NULL ORDER BY "payments"\."id" LIMIT \$2`).
		WithArgs("CAP-123", 1). // ID and LIMIT
		WillReturnRows(sqlmock.NewRows([]string{"id", "order_id", "paypal_capture_id"}).
			AddRow(1, 10, "CAP-123"))

	payment, err := repo.FindByCaptureID(context.Background(), "CAP-123")

	assert.NoError(t, err)
	assert.NotNil(t, payment)
	assert.Equal(t, 1, payment.ID)
	assert.NotNil(t, payment.OrderID)
	assert.Equal(t, 10, *payment.OrderID)
	assert.NotNil(t, payment.PaypalCaptureID)
	assert.Equal(t, "CAP-123", *payment.PaypalCaptureID)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
