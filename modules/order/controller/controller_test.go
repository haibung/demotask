package controller_test

import (
	"context"
	"testing"

	"github.com/demotask/backend/config"
	"github.com/demotask/backend/enum"
	"github.com/demotask/backend/models"
	"github.com/demotask/backend/modules/order"
	"github.com/demotask/backend/modules/order/controller"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/paypal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockOrderRepository
type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) CreateOrder(ctx context.Context, o *models.Order, tx *gorm.DB) (*int, error) {
	args := m.Called(ctx, o, tx)
	if args.Get(0) != nil {
		id := args.Get(0).(int)
		return &id, args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockOrderRepository) CreateOrderItems(ctx context.Context, items []models.OrderItem, tx *gorm.DB) error {
	args := m.Called(ctx, items, tx)
	return args.Error(0)
}
func (m *MockOrderRepository) FindByID(ctx context.Context, id int) (*models.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Order), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockOrderRepository) FindOrderID(ctx context.Context, req *models.Order) (*models.Order, error) {
	args := m.Called(ctx, req)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Order), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockOrderRepository) FindByPaypalOrderID(ctx context.Context, o *models.Order) (*models.Order, error) {
	args := m.Called(ctx, o)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Order), args.Error(1)
	}
	return nil, args.Error(1)
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
	args := m.Called(ctx, paypalOrderID)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Order), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockOrderRepository) FindByIDWithPayment(ctx context.Context, orderID int) (*models.Order, *models.Payment, error) {
	args := m.Called(ctx, orderID)
	var o *models.Order
	var p *models.Payment
	if args.Get(0) != nil {
		o = args.Get(0).(*models.Order)
	}
	if args.Get(1) != nil {
		p = args.Get(1).(*models.Payment)
	}
	return o, p, args.Error(2)
}
func (m *MockOrderRepository) FindByUserID(ctx context.Context, userID int) ([]models.Order, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.Order), args.Error(1)
}
func (m *MockOrderRepository) FindAllWithRelationsByUserID(ctx context.Context, userID int) ([]models.Order, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.Order), args.Error(1)
}

// MockProductRepository
type MockProductRepository struct{ mock.Mock }

func (m *MockProductRepository) Create(ctx context.Context, reqData *models.Product, tx *gorm.DB) (*int, error) {
	return nil, nil
}
func (m *MockProductRepository) Find(ctx context.Context, reqData *models.Product) (*models.Product, error) {
	return nil, nil
}
func (m *MockProductRepository) CreateOneTimePrice(ctx context.Context, reqData *models.ProductOneTimePrice, tx *gorm.DB) (*int, error) {
	return nil, nil
}
func (m *MockProductRepository) FindAll(ctx context.Context, filter *models.Product) ([]models.Product, error) {
	return nil, nil
}
func (m *MockProductRepository) FindByIDs(ctx context.Context, ids []int) ([]models.Product, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]models.Product), args.Error(1)
}
func (m *MockProductRepository) UpdatePlanID(ctx context.Context, productID int64, planID string, tx *gorm.DB) error {
	return nil
}
func (m *MockProductRepository) FindAllWithPrice(ctx context.Context) ([]models.Product, error) {
	return nil, nil
}

// MockPaymentRepository
type MockPaymentRepository struct{ mock.Mock }

func (m *MockPaymentRepository) CreateVaultToken(ctx context.Context, req *models.VaultToken, tx *gorm.DB) error {
	return nil
}
func (m *MockPaymentRepository) GetVaultTokenByUserID(ctx context.Context, userID int) (*models.VaultToken, error) {
	return nil, nil
}
func (m *MockPaymentRepository) GetVaultTokenByID(ctx context.Context, id int) (*models.VaultToken, error) {
	return nil, nil
}
func (m *MockPaymentRepository) FindByOrderID(ctx context.Context, orderID int) (*models.Payment, error) {
	return nil, nil
}
func (m *MockPaymentRepository) Create(ctx context.Context, payment *models.Payment, tx *gorm.DB) (*int, error) {
	return nil, nil
}
func (m *MockPaymentRepository) Update(ctx context.Context, payment *models.Payment, tx *gorm.DB) error {
	return nil
}
func (m *MockPaymentRepository) FindPendingByOrderID(ctx context.Context, orderID int) (*models.Payment, error) {
	return nil, nil
}
func (m *MockPaymentRepository) FindByCaptureID(ctx context.Context, captureID string) (*models.Payment, error) {
	return nil, nil
}
func (m *MockPaymentRepository) FindByID(ctx context.Context, arg1 int, arg2 int) (*models.VaultToken, error) {
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
	args := m.Called(req)
	if args.Get(0) != nil {
		return args.Get(0).(*paypal.OrderResponse), args.Error(1)
	}
	return nil, args.Error(1)
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
	return nil, nil
}
func (m *MockPayPalClient) VerifyWebhookSignature(req *paypal.WebhookVerificationRequest) (bool, error) {
	return false, nil
}
func (m *MockPayPalClient) GetWebhookID() string               { return "" }
func (m *MockPayPalClient) DeactivatePlan(planID string) error { return nil }
func (m *MockPayPalClient) GetSubscriptionDetail(subscriptionID string) (*paypal.SubscriptionDetailResponse, error) {
	return nil, nil
}

func TestCreateOrder(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockProductRepo := new(MockProductRepository)
	mockPaymentRepo := new(MockPaymentRepository)
	mockPayPal := new(MockPayPalClient)
	log := logger.NewLogger()
	cfg := &config.Config{}

	c := controller.NewController(controller.OrderController{
		Config:       cfg,
		Logger:       log,
		OrderRepo:    mockOrderRepo,
		ProductRepo:  mockProductRepo,
		PaymentRepo:  mockPaymentRepo,
		PayPalClient: mockPayPal,
	})

	ctx := context.Background()
	userID := 1

	req := &order.CreateOrderRequest{
		ContextUserID: &userID,
		Items: []order.CreateOrderItemRequest{
			{ProductID: 1, Quantity: 2},
		},
	}

	products := []models.Product{
		{
			ID:   1,
			Name: "AI Agent",
			OneTimePrices: []models.ProductOneTimePrice{
				{Price: 10.0, Currency: "USD"},
			},
		},
	}

	mockProductRepo.On("FindByIDs", ctx, []int{1}).Return(products, nil)

	mockOrderRepo.On("CreateOrder", ctx, mock.AnythingOfType("*models.Order"), (*gorm.DB)(nil)).Return(1, nil)
	mockOrderRepo.On("CreateOrderItems", ctx, mock.AnythingOfType("[]models.OrderItem"), (*gorm.DB)(nil)).Return(nil)

	resp, err := c.CreateOrder(ctx, req, nil)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "1", resp.OrderID)

	mockOrderRepo.AssertExpectations(t)
	mockProductRepo.AssertExpectations(t)
}
