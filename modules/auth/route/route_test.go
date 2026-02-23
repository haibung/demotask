package route_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/demotask/backend/modules/auth"
	"github.com/demotask/backend/modules/auth/route"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/postgres"
	"github.com/demotask/backend/routers"
	"github.com/demotask/backend/utilities"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// MockAuthController
type MockAuthController struct{ mock.Mock }

func (m *MockAuthController) Login(ctx context.Context, reqData *auth.LoginRequest) (*auth.LoginResponse, error) {
	args := m.Called(ctx, reqData)
	if args.Get(0) != nil {
		return args.Get(0).(*auth.LoginResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthController) Register(ctx context.Context, reqData *auth.RegisterRequest, tx *gorm.DB) error {
	args := m.Called(ctx, reqData, tx)
	return args.Error(0)
}

func setupTestDB(t *testing.T) (*postgres.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(gormPostgres.New(gormPostgres.Config{
		Conn: db,
	}), &gorm.Config{})
	assert.NoError(t, err)

	return &postgres.DB{
		Sql:  db,
		Gorm: gormDB,
	}, mock
}

type setupProps struct {
	echoApp   *echo.Echo
	mockCtrl  *MockAuthController
	mockDBReq sqlmock.Sqlmock
}

func setupRouteTest(t *testing.T) setupProps {

	e := echo.New()
	log := logger.NewLogger()
	routerWrapper := &routers.Router{Echo: e}
	db, mockDBReq := setupTestDB(t)
	mockCtrl := new(MockAuthController)

	handler := route.Handler{
		Controller: mockCtrl,
		Logger:     log,
		DB:         db,
		Router:     routerWrapper,
	}

	route.NewRoute(handler)

	return setupProps{
		echoApp:   e,
		mockCtrl:  mockCtrl,
		mockDBReq: mockDBReq,
	}
}

func TestAuthRoute_Login_Success(t *testing.T) {
	props := setupRouteTest(t)

	reqPayload := auth.LoginRequest{
		Email:    "test@test.com",
		Password: "Password123",
	}
	body, _ := json.Marshal(reqPayload)

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	props.mockDBReq.ExpectBegin()
	props.mockDBReq.ExpectCommit()

	props.mockCtrl.On("Login", mock.Anything, mock.AnythingOfType("*auth.LoginRequest")).Return(&auth.LoginResponse{
		AccessToken:  "access123",
		RefreshToken: "refresh123",
	}, nil)

	props.echoApp.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var res map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &res)
	assert.Equal(t, utilities.Success, res["message"])

	data := res["data"].(map[string]interface{})
	assert.Equal(t, "access123", data["access_token"])

	props.mockCtrl.AssertExpectations(t)
}

func TestAuthRoute_Register_Success(t *testing.T) {
	props := setupRouteTest(t)

	reqPayload := auth.RegisterRequest{
		Email:    "test@test.com",
		Password: "Password123",
		FullName: "Test User",
	}
	body, _ := json.Marshal(reqPayload)

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	props.mockDBReq.ExpectBegin()
	props.mockDBReq.ExpectCommit()

	props.mockCtrl.On("Register", mock.Anything, mock.AnythingOfType("*auth.RegisterRequest"), mock.AnythingOfType("*gorm.DB")).Return(nil)

	props.echoApp.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var res map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &res)
	assert.Equal(t, utilities.Success, res["message"])

	props.mockCtrl.AssertExpectations(t)
}
