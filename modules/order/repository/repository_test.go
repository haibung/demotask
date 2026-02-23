package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/demotask/backend/enum"
	"github.com/demotask/backend/models"
	"github.com/demotask/backend/modules/order/repository"
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

func setupRepository(t *testing.T) (repository.IOrderRepository, sqlmock.Sqlmock) {
	db, mock := setupTestDB(t)
	log := logger.NewLogger()

	repo := repository.NewRepository(repository.OrderRepository{
		Logger: log,
		DB:     db,
	})

	return repo, mock
}

func TestCreateOrder(t *testing.T) {
	repo, mock := setupRepository(t)

	now := time.Now()
	order := &models.Order{
		UserID:    1,
		Status:    enum.OrderStatusInternalPending,
		OrderType: enum.OrderTypeOneTime,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`^INSERT INTO "orders"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	id, err := repo.CreateOrder(context.Background(), order, nil)
	assert.NoError(t, err)
	if id != nil {
		assert.Equal(t, 1, *id)
	}

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestFindByID(t *testing.T) {
	repo, mock := setupRepository(t)

	mock.ExpectQuery(`^SELECT \* FROM "orders" WHERE id = \$1 AND "orders"\."deleted_at" IS NULL ORDER BY "orders"\."id" LIMIT \$2`).
		WithArgs(1, 1). // ID and LIMIT
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}).
			AddRow(1, 2))

	mock.ExpectQuery(`^SELECT \* FROM "order_items" WHERE "order_items"\."order_id" = \$1 AND "order_items"\."deleted_at" IS NULL`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "order_id", "product_name"}).
			AddRow(1, 1, "Product 1"))

	order, err := repo.FindByID(context.Background(), 1)

	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, 1, order.ID)
	assert.Equal(t, 2, order.UserID)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
