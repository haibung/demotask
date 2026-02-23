package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/demotask/backend/models"
	"github.com/demotask/backend/modules/product/repository"
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

func setupRepository(t *testing.T) (repository.IProductRepository, sqlmock.Sqlmock) {
	db, mock := setupTestDB(t)
	log := logger.NewLogger()

	repo := repository.NewRepository(repository.ProductRepository{
		Logger: log,
		DB:     db,
	})

	return repo, mock
}

func TestCreateProduct(t *testing.T) {
	repo, mock := setupRepository(t)

	now := time.Now()
	product := &models.Product{
		Name:            "AI Agent",
		Description:     "Test Description",
		BusinessType:    "DIGITAL",
		PaypalProductID: "PROD-123",
		Category:        "SOFTWARE",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`^INSERT INTO "products"`).
		WithArgs(
			product.PaypalProductID, product.Name, product.Description,
			product.BusinessType, product.Category, product.ImageURL,
			product.HomeURL, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	id, err := repo.Create(context.Background(), product, nil)
	assert.NoError(t, err)
	if id != nil {
		assert.Equal(t, 1, *id)
	}

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestFindProduct(t *testing.T) {
	repo, mock := setupRepository(t)

	mock.ExpectQuery(`^SELECT \* FROM "products" WHERE "products"\."id" = \$1 AND "products"\."deleted_at" IS NULL ORDER BY "products"\."id" LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description"}).
			AddRow(1, "AI Agent", "Test Description"))

	reqData := &models.Product{ID: 1}
	product, err := repo.Find(context.Background(), reqData)

	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, 1, product.ID)
	assert.Equal(t, "AI Agent", product.Name)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
