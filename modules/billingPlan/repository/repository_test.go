package repository_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/demotask/backend/enum"
	"github.com/demotask/backend/models"
	"github.com/demotask/backend/modules/billingPlan/repository"
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

func setupRepository(t *testing.T) (repository.IBillingPlanRepository, sqlmock.Sqlmock) {
	db, mock := setupTestDB(t)
	log := logger.NewLogger()

	repo := repository.NewRepository(repository.BillingPlanRepository{
		Logger: log,
		DB:     db,
	})

	return repo, mock
}

func TestCreatePlan(t *testing.T) {
	repo, mock := setupRepository(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`^INSERT INTO "billing_plans"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	desc := "Test Plan"
	plan := &models.BillingPlan{
		ProductID:   1,
		Name:        "Test Plan Name",
		Description: &desc,
		Status:      enum.PlanStatusActive,
	}

	id, err := repo.Create(context.Background(), plan, nil)
	assert.NoError(t, err)
	if id != nil {
		assert.Equal(t, 1, *id)
	}

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestFindByID(t *testing.T) {
	repo, mock := setupRepository(t)

	mock.ExpectQuery(`^SELECT \* FROM "billing_plans" WHERE id = \$1 AND "billing_plans"\."deleted_at" IS NULL ORDER BY "billing_plans"\."id" LIMIT \$2`).
		WithArgs(1, 1). // ID and LIMIT
		WillReturnRows(sqlmock.NewRows([]string{"id", "product_id"}).
			AddRow(1, 10))

	mock.ExpectQuery(`^SELECT \* FROM "billing_cycles" WHERE "billing_cycles"\."billing_plan_id" = \$1 AND "billing_cycles"\."deleted_at" IS NULL`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "billing_plan_id", "interval_unit"}).
			AddRow(1, 1, "MONTH"))

	plan, err := repo.FindByID(context.Background(), &models.BillingPlan{ID: 1})

	assert.NoError(t, err)
	assert.NotNil(t, plan)
	assert.Equal(t, 1, plan.ID)
	assert.Equal(t, 10, plan.ProductID)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
