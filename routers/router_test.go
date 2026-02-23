package routers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/demotask/backend/models"
	"github.com/demotask/backend/modules/user"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/routers"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockUserRepository
type MockUserRepository struct{ mock.Mock }

func (m *MockUserRepository) Create(ctx context.Context, reqData *models.User, tx *gorm.DB) (*int, error) {
	args := m.Called(ctx, reqData, tx)
	if args.Get(0) != nil {
		id := args.Get(0).(int)
		return &id, args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockUserRepository) FindByID(ctx context.Context, reqData *models.User) (*models.User, error) {
	args := m.Called(ctx, reqData)
	if args.Get(0) != nil {
		return args.Get(0).(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockUserRepository) FindByEmail(ctx context.Context, reqData *models.User) (*models.User, error) {
	args := m.Called(ctx, reqData)
	if args.Get(0) != nil {
		return args.Get(0).(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func setupTestRouter() (*echo.Echo, *MockUserRepository) {
	mockRepo := new(MockUserRepository)
	log := logger.NewLogger()
	routerApp := routers.NewRouter(mockRepo, log)

	routerApp.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	routerApp.POST("/api/v1/user", func(c echo.Context) error {
		var req user.CreateRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid format"})
		}

		if req.Email == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "email required"})
		}

		_, err := mockRepo.Create(c.Request().Context(), &models.User{Email: req.Email}, nil)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal error"})
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "created"})
	})

	return routerApp.Echo, mockRepo
}

func TestSmoke_HealthCheck(t *testing.T) {
	e, _ := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
}

func TestRegression_MalformedJSONBody(t *testing.T) {
	e, _ := setupTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user", strings.NewReader(`{invalid_json`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRegression_EmptyEmail(t *testing.T) {
	e, _ := setupTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user", strings.NewReader(`{"email": ""}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), `"error":"email required"`)
}

func TestSecurity_RateLimiter(t *testing.T) {
	e, _ := setupTestRouter()

	// Spin up a real local test server to ensure IP and rate limit middleware works properly
	ts := httptest.NewServer(e)
	defer ts.Close()

	client := &http.Client{}

	// Rate limit is 5 rps with a burst of 15. The capacity is burst (15).
	// To reliably hit the limit we need >15 requests rapidly.
	for i := 0; i < 25; i++ {
		req, _ := http.NewRequest(http.MethodGet, ts.URL+"/health", nil)
		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			// Successfully triggered rate limit
			return
		}
	}
	t.Errorf("Rate limiter did not block requests")
}
