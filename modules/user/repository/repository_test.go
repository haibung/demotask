package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/demotask/backend/models"
	"github.com/demotask/backend/modules/user/repository"
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

func setupRepository(t *testing.T) (repository.IUserInterface, sqlmock.Sqlmock) {
	db, mock := setupTestDB(t)
	log := logger.NewLogger()

	repo := repository.NewRepository(repository.UserRepository{
		DB:     db,
		Logger: log,
	})

	return repo, mock
}

func TestCreateUser(t *testing.T) {
	repo, mock := setupRepository(t)

	now := time.Now()
	user := &models.User{
		FullName:  "Test User",
		Email:     "test@test.com",
		Password:  "password123",
		CreatedAt: now,
		UpdatedAt: now,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`^INSERT INTO "users"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	id, err := repo.Create(context.Background(), user, nil)
	assert.NoError(t, err)
	if id != nil {
		assert.Equal(t, 1, *id)
	}

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestFindByID(t *testing.T) {
	repo, mock := setupRepository(t)

	mock.ExpectQuery(`^SELECT \* FROM "users" WHERE "users"\."id" = \$1 AND "users"\."deleted_at" IS NULL ORDER BY "users"\."id" LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "full_name", "email"}).
			AddRow(1, "Test User", "test@test.com"))

	user, err := repo.FindByID(context.Background(), &models.User{ID: 1})

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, 1, user.ID)
	assert.Equal(t, "Test User", user.FullName)
	assert.Equal(t, "test@test.com", user.Email)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestFindByEmail(t *testing.T) {
	repo, mock := setupRepository(t)

	mock.ExpectQuery(`^SELECT \* FROM "users" WHERE "users"\."email" = \$1 AND "users"\."deleted_at" IS NULL ORDER BY "users"\."id" LIMIT \$2`).
		WithArgs("test@test.com", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "full_name", "email"}).
			AddRow(1, "Test User", "test@test.com"))

	user, err := repo.FindByEmail(context.Background(), &models.User{Email: "test@test.com"})

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, 1, user.ID)
	assert.Equal(t, "test@test.com", user.Email)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
