package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/demotask/backend/enum"
	"github.com/demotask/backend/models"
	"github.com/demotask/backend/modules/subscription/repository"
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

func setupRepository(t *testing.T) (repository.ISubscriptionRepository, sqlmock.Sqlmock) {
	db, mock := setupTestDB(t)
	log := logger.NewLogger()

	repo := repository.NewRepository(repository.SubscriptionRepository{
		DB:     db,
		Logger: log,
	})

	return repo, mock
}

func TestCreateSubscription(t *testing.T) {
	repo, mock := setupRepository(t)

	now := time.Now()
	startTime := now.Add(time.Hour)
	subscription := &models.Subscription{
		UserID:               1,
		OrderID:              2,
		BillingPlanID:        3,
		PaypalSubscriptionID: "SUB-123",
		Status:               enum.SubscriptionStatusActive,
		StartTime:            &startTime,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`^INSERT INTO "subscriptions"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	id, err := repo.Create(context.Background(), subscription, nil)
	assert.NoError(t, err)
	if id != nil {
		assert.Equal(t, 1, *id)
	}

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestFindByPayPalID(t *testing.T) {
	repo, mock := setupRepository(t)

	mock.ExpectQuery(`^SELECT \* FROM "subscriptions" WHERE paypal_subscription_id = \$1 AND "subscriptions"\."deleted_at" IS NULL ORDER BY "subscriptions"\."id" LIMIT \$2`).
		WithArgs("SUB-123", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "paypal_subscription_id"}).
			AddRow(1, 10, "SUB-123"))

	sub, err := repo.FindByPayPalID(context.Background(), "SUB-123")

	assert.NoError(t, err)
	assert.NotNil(t, sub)
	assert.Equal(t, 1, sub.ID)
	assert.Equal(t, 10, sub.UserID)
	assert.Equal(t, "SUB-123", sub.PaypalSubscriptionID)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
