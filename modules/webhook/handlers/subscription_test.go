package handlers_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/demotask/backend/models"
	"github.com/demotask/backend/modules/webhook/handlers"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/paypal"
	"github.com/demotask/backend/packages/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

// MockPaymentRepository
type MockPaymentRepository struct{ mock.Mock }

func (m *MockPaymentRepository) Create(ctx context.Context, payment *models.Payment, tx *gorm.DB) (*int, error) {
	args := m.Called(ctx, payment, tx)
	if args.Get(0) != nil {
		return args.Get(0).(*int), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockPaymentRepository) FindByCaptureID(ctx context.Context, captureID string) (*models.Payment, error) {
	args := m.Called(ctx, captureID)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Payment), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockPaymentRepository) FindByID(ctx context.Context, id int, userID int) (*models.VaultToken, error) {
	return nil, nil
}
func (m *MockPaymentRepository) FindVaultByUserID(ctx context.Context, userID int) ([]models.VaultToken, error) {
	return nil, nil
}

// MockPayPalClient
type MockPayPalClient struct{ mock.Mock }

func (m *MockPayPalClient) GetAuthorizationURL(state string) string { return "" }
func (m *MockPayPalClient) ExchangeCodeForToken(code string) (*paypal.TokenResponse, error) {
	return nil, nil
}
func (m *MockPayPalClient) RefreshAccessToken(refreshToken string) (*paypal.TokenResponse, error) {
	return nil, nil
}
func (m *MockPayPalClient) ClientCredentials() (*paypal.TokenResponse, error)        { return nil, nil }
func (m *MockPayPalClient) GetUserInfo(accessToken string) (*paypal.UserInfo, error) { return nil, nil }
func (m *MockPayPalClient) RevokeToken(token string) error                           { return nil }
func (m *MockPayPalClient) CreateOrder(req *paypal.OrderRequest) (*paypal.OrderResponse, error) {
	return nil, nil
}
func (m *MockPayPalClient) CaptureOrder(orderID string) (*paypal.CaptureResponse, error) {
	args := m.Called(orderID)
	if args.Get(0) != nil {
		return args.Get(0).(*paypal.CaptureResponse), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockPayPalClient) CreateSubscription(req *paypal.CreateSubscriptionRequest) (*paypal.CreateSubscriptionResponse, error) {
	return nil, nil
}
func (m *MockPayPalClient) CreateProduct(req *paypal.CreateProductRequest) (*paypal.CreateProductResponse, error) {
	return nil, nil
}
func (m *MockPayPalClient) CreatePlan(req *paypal.CreatePlanRequest, requestID string) (*paypal.CreatePlanResponse, error) {
	return nil, nil
}
func (m *MockPayPalClient) CreateVaultSetupToken(req *paypal.VaultSetupTokenRequest) (*paypal.VaultSetupTokenResponse, error) {
	return nil, nil
}
func (m *MockPayPalClient) CreatePaymentToken(req *paypal.CreatePaymentTokenRequest) (*paypal.CreatePaymentTokenResponse, error) {
	return nil, nil
}
func (m *MockPayPalClient) CreateOrderWithVault(amount float64, currency string, vaultID string) (*paypal.CreateOrderResponse, error) {
	return nil, nil
}
func (m *MockPayPalClient) VerifyWebhookSignature(req *paypal.WebhookVerificationRequest) (bool, error) {
	return false, nil
}
func (m *MockPayPalClient) GetWebhookID() string               { return "" }
func (m *MockPayPalClient) DeactivatePlan(planID string) error { return nil }
func (m *MockPayPalClient) GetSubscriptionDetail(subscriptionID string) (*paypal.SubscriptionDetailResponse, error) {
	args := m.Called(subscriptionID)
	if args.Get(0) != nil {
		return args.Get(0).(*paypal.SubscriptionDetailResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestHandleSubscriptionCreated(t *testing.T) {
	db, sqlMock := setupTestDB(t)
	mockPayPal := new(MockPayPalClient)
	mockPayment := new(MockPaymentRepository)
	log := logger.NewLogger()

	handler := handlers.NewSubscriptionHandler(handlers.SubscriptionHandler{
		DB:           db,
		Logger:       log,
		PayPalClient: mockPayPal,
		PaymentRepo:  mockPayment,
	})

	ctx := context.Background()
	payload := map[string]interface{}{
		"resource": map[string]interface{}{
			"id": "SUB-123",
		},
	}

	sqlMock.ExpectQuery(`^SELECT \* FROM "subscriptions" WHERE paypal_subscription_id = \$1 AND "subscriptions"\."deleted_at" IS NULL ORDER BY "subscriptions"\."id" LIMIT \$2`).
		WithArgs("SUB-123", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "order_id"}).AddRow(1, "", 10))

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(`^UPDATE "subscriptions" SET "status"=\$1,"updated_at"=\$2 WHERE paypal_subscription_id = \$3`).
		WithArgs("APPROVAL_PENDING", sqlmock.AnyArg(), "SUB-123").
		WillReturnResult(sqlmock.NewResult(1, 1))
	sqlMock.ExpectCommit()

	err := handler.HandleSubscription(ctx, "BILLING.SUBSCRIPTION.CREATED", payload, db.Gorm)

	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestHandleSale(t *testing.T) {
	db, sqlMock := setupTestDB(t)
	mockPayPal := new(MockPayPalClient)
	mockPayment := new(MockPaymentRepository)
	log := logger.NewLogger()

	handler := handlers.NewSubscriptionHandler(handlers.SubscriptionHandler{
		DB:           db,
		Logger:       log,
		PayPalClient: mockPayPal,
		PaymentRepo:  mockPayment,
	})

	ctx := context.Background()
	now := time.Now().UTC()
	payload := map[string]interface{}{
		"resource": map[string]interface{}{
			"id":                   "SALE-123",
			"billing_agreement_id": "SUB-123",
			"amount": map[string]interface{}{
				"total":    "10.00",
				"currency": "USD",
			},
			"create_time": now.Format(time.RFC3339),
		},
	}

	mockPayment.On("FindByCaptureID", ctx, "SALE-123").Return((*models.Payment)(nil), nil)

	sqlMock.ExpectQuery(`^SELECT \* FROM "subscriptions" WHERE paypal_subscription_id = \$1 AND "subscriptions"\."deleted_at" IS NULL ORDER BY "subscriptions"\."id" LIMIT \$2`).
		WithArgs("SUB-123", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}).AddRow(1, 10))

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(`^INSERT INTO "payments"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	sqlMock.ExpectCommit()

	mockPayPal.On("GetSubscriptionDetail", "SUB-123").Return(&paypal.SubscriptionDetailResponse{}, nil)

	err := handler.HandleSale(ctx, payload, db.Gorm)

	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	mockPayment.AssertExpectations(t)
	mockPayPal.AssertExpectations(t)
}
