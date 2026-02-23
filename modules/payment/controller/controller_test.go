package controller_test

import (
	"context"
	"testing"

	"github.com/demotask/backend/enum"
	"github.com/demotask/backend/models"
	"github.com/demotask/backend/modules/payment"
	"github.com/demotask/backend/modules/payment/controller"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/paypal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockPaymentRepository
type MockPaymentRepository struct{ mock.Mock }

func (m *MockPaymentRepository) Create(ctx context.Context, payment *models.Payment, tx *gorm.DB) (*int, error) {
	return nil, nil
}
func (m *MockPaymentRepository) FindByCaptureID(ctx context.Context, captureID string) (*models.Payment, error) {
	return nil, nil
}
func (m *MockPaymentRepository) FindByID(ctx context.Context, id int, userID int) (*models.VaultToken, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) != nil {
		return args.Get(0).(*models.VaultToken), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockPaymentRepository) FindVaultByUserID(ctx context.Context, userID int) ([]models.VaultToken, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.VaultToken), args.Error(1)
}

// MockOrderRepository
type MockOrderRepository struct{ mock.Mock }

func (m *MockOrderRepository) CreateOrder(ctx context.Context, o *models.Order, tx *gorm.DB) (*int, error) {
	return nil, nil
}
func (m *MockOrderRepository) CreateOrderItems(ctx context.Context, items []models.OrderItem, tx *gorm.DB) error {
	return nil
}
func (m *MockOrderRepository) FindByID(ctx context.Context, id int) (*models.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Order), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockOrderRepository) FindOrderID(ctx context.Context, req *models.Order) (*models.Order, error) {
	return nil, nil
}
func (m *MockOrderRepository) FindByPaypalOrderID(ctx context.Context, o *models.Order) (*models.Order, error) {
	return nil, nil
}
func (m *MockOrderRepository) UpdatePaypalOrderID(ctx context.Context, orderID int, paypalOrderID string, tx *gorm.DB) error {
	args := m.Called(ctx, orderID, paypalOrderID, tx)
	return args.Error(0)
}
func (m *MockOrderRepository) UpdateStatus(ctx context.Context, orderID int, status enum.OrderStatusInternal, tx *gorm.DB) error {
	args := m.Called(ctx, orderID, status, tx)
	return args.Error(0)
}
func (m *MockOrderRepository) FindOrderPaypalID(ctx context.Context, paypalOrderID string) (*models.Order, error) {
	return nil, nil
}
func (m *MockOrderRepository) FindByIDWithPayment(ctx context.Context, orderID int) (*models.Order, *models.Payment, error) {
	return nil, nil, nil
}
func (m *MockOrderRepository) FindByUserID(ctx context.Context, userID int) ([]models.Order, error) {
	return nil, nil
}
func (m *MockOrderRepository) FindAllWithRelationsByUserID(ctx context.Context, userID int) ([]models.Order, error) {
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
	return nil, nil
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
	args := m.Called(amount, currency, vaultID)
	if args.Get(0) != nil {
		return args.Get(0).(*paypal.CreateOrderResponse), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockPayPalClient) VerifyWebhookSignature(req *paypal.WebhookVerificationRequest) (bool, error) {
	return false, nil
}
func (m *MockPayPalClient) GetWebhookID() string               { return "" }
func (m *MockPayPalClient) DeactivatePlan(planID string) error { return nil }
func (m *MockPayPalClient) GetSubscriptionDetail(subscriptionID string) (*paypal.SubscriptionDetailResponse, error) {
	return nil, nil
}

func TestPayAgain(t *testing.T) {
	mockPaymentRepo := new(MockPaymentRepository)
	mockOrderRepo := new(MockOrderRepository)
	mockPayPal := new(MockPayPalClient)
	log := logger.NewLogger()

	c := controller.NewController(controller.PaymentController{
		Logger:       log,
		PaymentRepo:  mockPaymentRepo,
		OrderRepo:    mockOrderRepo,
		PayPalClient: mockPayPal,
	})

	ctx := context.Background()
	userID := 1
	orderID := 1
	vaultID := 1

	req := &payment.PayAgainRequest{
		ContextUserID: &userID,
		OrderID:       orderID,
		VaultTokenID:  vaultID,
	}

	order := &models.Order{
		ID:     orderID,
		UserID: userID,
		Status: enum.OrderStatusInternalPending,
		Items: []models.OrderItem{
			{TotalPrice: 50.0, Currency: "USD"},
		},
	}
	mockOrderRepo.On("FindByID", ctx, orderID).Return(order, nil)

	vault := &models.VaultToken{
		ID:            vaultID,
		UserID:        userID,
		PayPalVaultID: "VAULT-123",
	}
	mockPaymentRepo.On("FindByID", ctx, vaultID, userID).Return(vault, nil)

	mockPayPal.On("CreateOrderWithVault", float64(50.0), "USD", "VAULT-123").Return(&paypal.CreateOrderResponse{ID: "ORDER-123"}, nil)
	mockOrderRepo.On("UpdatePaypalOrderID", ctx, orderID, "ORDER-123", (*gorm.DB)(nil)).Return(nil)
	mockOrderRepo.On("UpdateStatus", ctx, orderID, enum.OrderStatusInternalWaitingPayment, (*gorm.DB)(nil)).Return(nil)

	resp, err := c.PayAgain(ctx, req, nil)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, orderID, resp.OrderID)
	assert.Equal(t, "ORDER-123", resp.PaypalOrderID)

	mockOrderRepo.AssertExpectations(t)
	mockPaymentRepo.AssertExpectations(t)
	mockPayPal.AssertExpectations(t)
}

func TestGetVaultTokens(t *testing.T) {
	mockPaymentRepo := new(MockPaymentRepository)
	c := controller.NewController(controller.PaymentController{
		PaymentRepo: mockPaymentRepo,
	})

	ctx := context.Background()
	userID := 1
	req := &payment.GetVaultTokensRequest{ContextUserID: &userID}

	tokens := []models.VaultToken{
		{ID: 1, UserID: userID, IsDefault: true},
	}
	mockPaymentRepo.On("FindVaultByUserID", ctx, userID).Return(tokens, nil)

	resp, err := c.GetVaultTokens(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Items, 1)
	assert.Equal(t, 1, resp.Items[0].ID)

	mockPaymentRepo.AssertExpectations(t)
}
