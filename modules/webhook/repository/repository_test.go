package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/demotask/backend/enum"
	"github.com/demotask/backend/models"
	"github.com/demotask/backend/modules/webhook/repository"
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

func setupRepository(t *testing.T) (repository.IWebhookRepository, sqlmock.Sqlmock) {
	db, mock := setupTestDB(t)
	log := logger.NewLogger()

	repo := repository.NewRepository(repository.WebhookRepository{
		DB:     db,
		Logger: log,
	})

	return repo, mock
}

func TestFindByEventID(t *testing.T) {
	repo, mock := setupRepository(t)

	mock.ExpectQuery(`^SELECT \* FROM "webhook_events" WHERE event_id = \$1 AND "webhook_events"\."deleted_at" IS NULL ORDER BY "webhook_events"\."id" LIMIT \$2`).
		WithArgs("EVT-123", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "event_id", "status"}).
			AddRow(1, "EVT-123", "PENDING"))

	event, err := repo.FindByEventID(context.Background(), "EVT-123")

	assert.NoError(t, err)
	assert.NotNil(t, event)
	assert.Equal(t, 1, event.ID)
	assert.Equal(t, "EVT-123", event.EventID)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestSave(t *testing.T) {
	repo, mock := setupRepository(t)

	now := time.Now()
	event := &models.WebhookEvent{
		Provider:  "paypal",
		EventID:   "EVT-123",
		EventType: "PAYMENT.CAPTURE.COMPLETED",
		Payload:   "{}",
		Status:    enum.WebhookStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`^INSERT INTO "webhook_events"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	id, err := repo.Save(context.Background(), event, repo.(*repository.WebhookRepository).DB.Gorm)

	assert.NoError(t, err)
	assert.NotNil(t, id)
	assert.Equal(t, 1, *id)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestMarkProcessed(t *testing.T) {
	repo, mock := setupRepository(t)

	mock.ExpectBegin()
	mock.ExpectExec(`^UPDATE "webhook_events" SET "processed_at"=\$1,"status"=\$2,"updated_at"=\$3 WHERE id = \$4 AND "webhook_events"\."deleted_at" IS NULL`).
		WithArgs(sqlmock.AnyArg(), enum.WebhookStatusProcessed.String(), sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.MarkProcessed(context.Background(), 1, repo.(*repository.WebhookRepository).DB.Gorm)

	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestMarkFailed(t *testing.T) {
	repo, mock := setupRepository(t)

	mock.ExpectBegin()
	mock.ExpectExec(`^UPDATE "webhook_events" SET "status"=\$1,"updated_at"=\$2 WHERE id = \$3 AND "webhook_events"\."deleted_at" IS NULL`).
		WithArgs(enum.WebhookStatusFailed.String(), sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.MarkFailed(context.Background(), 1, repo.(*repository.WebhookRepository).DB.Gorm)

	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
