package route_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/demotask/backend/modules/webhook/route"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/postgres"
	"github.com/demotask/backend/routers"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// MockWebhookController
type MockWebhookController struct{ mock.Mock }

func (m *MockWebhookController) HandlePayPalWebhook(ctx context.Context, headers map[string]string, rawBody []byte, body map[string]interface{}, tx *gorm.DB) error {
	args := m.Called(ctx, headers, rawBody, body, tx)
	return args.Error(0)
}

func setupTestDB(t *testing.T) (*postgres.DB, sqlmock.Sqlmock) {
	db, mockDBReq, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(gormPostgres.New(gormPostgres.Config{
		Conn: db,
	}), &gorm.Config{})
	assert.NoError(t, err)

	return &postgres.DB{
		Sql:  db,
		Gorm: gormDB,
	}, mockDBReq
}

func TestWebhookRoute_HandlePayPalWebhook_Success(t *testing.T) {
	e := echo.New()
	log := logger.NewLogger()
	routerWrapper := &routers.Router{Echo: e}
	db, mockDBReq := setupTestDB(t)
	mockCtrl := new(MockWebhookController)

	handler := route.Handler{
		Controller: mockCtrl,
		Logger:     log,
		DB:         db,
		Router:     routerWrapper,
	}

	route.NewRoute(handler)

	reqPayload := map[string]interface{}{
		"id":         "WH-123",
		"event_type": "PAYMENT.CAPTURE.COMPLETED",
	}
	body, _ := json.Marshal(reqPayload)

	req := httptest.NewRequest(http.MethodPost, "/v1/webhooks/paypal", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	// Set mock headers
	req.Header.Set("Paypal-Auth-Algo", "ALGO")
	req.Header.Set("Paypal-Cert-Url", "URL")
	req.Header.Set("Paypal-Transmission-Id", "TID")
	req.Header.Set("Paypal-Transmission-Sig", "SIG")
	req.Header.Set("Paypal-Transmission-Time", "TIME")
	req.Header.Set("Paypal-Webhook-Id", "WID")

	rec := httptest.NewRecorder()

	mockDBReq.ExpectBegin()
	mockDBReq.ExpectCommit()

	mockCtrl.On("HandlePayPalWebhook", mock.Anything, mock.AnythingOfType("map[string]string"), mock.AnythingOfType("[]uint8"), mock.AnythingOfType("map[string]interface {}"), mock.AnythingOfType("*gorm.DB")).Return(nil)

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	mockCtrl.AssertExpectations(t)
	assert.NoError(t, mockDBReq.ExpectationsWereMet())
}
