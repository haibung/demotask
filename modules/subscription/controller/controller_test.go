package controller_test

import (
	"context"
	"testing"
	"time"

	"github.com/demotask/backend/config"
	"github.com/demotask/backend/enum"
	"github.com/demotask/backend/models"
	"github.com/demotask/backend/modules/subscription"
	"github.com/demotask/backend/modules/subscription/controller"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/paypal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockSubscriptionRepository
type MockSubscriptionRepository struct{ mock.Mock }

func (m *MockSubscriptionRepository) Create(ctx context.Context, subscription *models.Subscription, tx *gorm.DB) (*int, error) {
	args := m.Called(ctx, subscription, tx)
	if args.Get(0) != nil {
		id := args.Get(0).(int)
		return &id, args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockSubscriptionRepository) FindByPayPalID(ctx context.Context, paypalSubscriptionID string) (*models.Subscription, error) {
	args := m.Called(ctx, paypalSubscriptionID)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Subscription), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockSubscriptionRepository) UpdateStatus(ctx context.Context, paypalSubscriptionID string, status string, tx *gorm.DB) error {
	args := m.Called(ctx, paypalSubscriptionID, status, tx)
	return args.Error(0)
}
func (m *MockSubscriptionRepository) UpdateStartTime(ctx context.Context, paypalSubscriptionID string, tx *gorm.DB) error {
	args := m.Called(ctx, paypalSubscriptionID, tx)
	return args.Error(0)
}
func (m *MockSubscriptionRepository) FindByID(ctx context.Context, id int) (*models.Subscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Subscription), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockSubscriptionRepository) FindByUserID(ctx context.Context, userID int) ([]models.Subscription, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.Subscription), args.Error(1)
}
func (m *MockSubscriptionRepository) FindByOrderID(ctx context.Context, orderID int) (*models.Subscription, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Subscription), args.Error(1)
	}
	return nil, args.Error(1)
}

// Mocks inherited from other tests
type MockBillingPlanRepository struct{ mock.Mock }
type MockOrderRepository struct{ mock.Mock }
type MockProductRepository struct{ mock.Mock }
type MockPayPalClient struct{ mock.Mock }
type MockPaymentRepository struct{ mock.Mock }

func (m *MockPaymentRepository) Create(ctx context.Context, payment *models.Payment, tx *gorm.DB) (*int, error) {
	return nil, nil
}
func (m *MockPaymentRepository) FindByCaptureID(ctx context.Context, captureID string) (*models.Payment, error) {
	return nil, nil
}
func (m *MockPaymentRepository) FindByID(ctx context.Context, id int, userID int) (*models.VaultToken, error) {
	return nil, nil
}
func (m *MockPaymentRepository) FindVaultByUserID(ctx context.Context, userID int) ([]models.VaultToken, error) {
	return nil, nil
}

func (m *MockBillingPlanRepository) Create(ctx context.Context, req *models.BillingPlan, tx *gorm.DB) (*int, error) {
	return nil, nil
}
func (m *MockBillingPlanRepository) ExistsActivePlan(ctx context.Context, productID int, name string) (bool, error) {
	return false, nil
}
func (m *MockBillingPlanRepository) CreateBillingCycle(ctx context.Context, req *models.BillingCycle, tx *gorm.DB) error {
	return nil
}
func (m *MockBillingPlanRepository) FindByID(ctx context.Context, req *models.BillingPlan) (*models.BillingPlan, error) {
	args := m.Called(ctx, req)
	if args.Get(0) != nil {
		return args.Get(0).(*models.BillingPlan), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockBillingPlanRepository) FindByPayPalID(ctx context.Context, paypalSubscriptionID string) (*models.Subscription, error) {
	return nil, nil
}
func (m *MockBillingPlanRepository) UpdateStatus(ctx context.Context, paypalSubscriptionID string, status string, tx *gorm.DB) error {
	return nil
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
	return nil
}
func (m *MockOrderRepository) FindByID(ctx context.Context, id int) (*models.Order, error) {
	return nil, nil
}
func (m *MockOrderRepository) FindOrderID(ctx context.Context, req *models.Order) (*models.Order, error) {
	return nil, nil
}
func (m *MockOrderRepository) FindByPaypalOrderID(ctx context.Context, o *models.Order) (*models.Order, error) {
	return nil, nil
}
func (m *MockOrderRepository) UpdatePaypalOrderID(ctx context.Context, orderID int, paypalOrderID string, tx *gorm.DB) error {
	return nil
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

func (m *MockProductRepository) Create(ctx context.Context, req *models.Product, tx *gorm.DB) (*int, error) {
	return nil, nil
}
func (m *MockProductRepository) Find(ctx context.Context, req *models.Product) (*models.Product, error) {
	return nil, nil
}
func (m *MockProductRepository) CreateOneTimePrice(ctx context.Context, req *models.ProductOneTimePrice, tx *gorm.DB) (*int, error) {
	return nil, nil
}
func (m *MockProductRepository) FindAll(ctx context.Context, filter *models.Product) ([]models.Product, error) {
	return nil, nil
}
func (m *MockProductRepository) FindByIDs(ctx context.Context, ids []int) ([]models.Product, error) {
	return nil, nil
}
func (m *MockProductRepository) UpdatePlanID(ctx context.Context, productID int64, planID string, tx *gorm.DB) error {
	return nil
}
func (m *MockProductRepository) FindAllWithPrice(ctx context.Context) ([]models.Product, error) {
	return nil, nil
}

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
	args := m.Called(req)
	if args.Get(0) != nil {
		return args.Get(0).(*paypal.CreateSubscriptionResponse), args.Error(1)
	}
	return nil, args.Error(1)
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

func TestCreateSubscription(t *testing.T) {
	mockSubRepo := new(MockSubscriptionRepository)
	mockPlanRepo := new(MockBillingPlanRepository)
	mockOrderRepo := new(MockOrderRepository)
	mockPaymentRepo := new(MockPaymentRepository)
	mockPayPal := new(MockPayPalClient)
	log := logger.NewLogger()
	cfg := &config.Config{}

	c := controller.NewController(controller.SubscriptionController{
		Config:           cfg,
		Logger:           log,
		SubscriptionRepo: mockSubRepo,
		BillingPlanRepo:  mockPlanRepo,
		OrderRepo:        mockOrderRepo,
		PaymentRepo:      mockPaymentRepo,
		PayPalClient:     mockPayPal,
	})

	ctx := context.Background()
	userID := 1

	req := &subscription.CreateSubscriptionRequest{
		ContextUserID: &userID,
		BillingPlanID: 1,
		Quantity:      1,
	}

	ppPlanID := "PLAN-123"
	plan := &models.BillingPlan{
		ID:           1,
		PaypalPlanID: &ppPlanID,
		BillingCycles: []models.BillingCycle{
			{TenureType: "REGULAR", PriceValue: 10.0, Currency: "USD"},
		},
	}
	mockPlanRepo.On("FindByID", ctx, &models.BillingPlan{ID: 1}).Return(plan, nil)

	mockOrderRepo.On("CreateOrder", ctx, mock.AnythingOfType("*models.Order"), (*gorm.DB)(nil)).Return(1, nil)

	ppRes := &paypal.CreateSubscriptionResponse{
		ID:     "SUB-123",
		Status: "APPROVAL_PENDING",
		Links: []paypal.Link{
			{Rel: "approve", Href: "https://paypal.com/approve"},
		},
	}
	mockPayPal.On("CreateSubscription", mock.AnythingOfType("*paypal.CreateSubscriptionRequest")).Return(ppRes, nil)

	mockSubRepo.On("Create", ctx, mock.AnythingOfType("*models.Subscription"), (*gorm.DB)(nil)).Return(1, nil)
	mockOrderRepo.On("UpdateStatus", ctx, 1, enum.OrderStatusInternalWaitingPayment, (*gorm.DB)(nil)).Return(nil)

	resp, err := c.CreateSubscription(ctx, req, nil)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, resp.OrderID)
	assert.Equal(t, "SUB-123", resp.SubscriptionID)
	assert.Equal(t, "https://paypal.com/approve", resp.ApprovalURL)

	mockPlanRepo.AssertExpectations(t)
	mockOrderRepo.AssertExpectations(t)
	mockPayPal.AssertExpectations(t)
	mockSubRepo.AssertExpectations(t)
}

func TestGetMySubscriptions(t *testing.T) {
	mockSubRepo := new(MockSubscriptionRepository)
	c := controller.NewController(controller.SubscriptionController{
		SubscriptionRepo: mockSubRepo,
	})

	ctx := context.Background()
	userID := 1

	now := time.Now()
	subs := []models.Subscription{
		{
			ID:                   1,
			OrderID:              2,
			BillingPlanID:        3,
			PaypalSubscriptionID: "SUB-123",
			Status:               enum.SubscriptionStatusActive,
			StartTime:            &now,
			CreatedAt:            now,
		},
	}
	mockSubRepo.On("FindByUserID", ctx, userID).Return(subs, nil)

	resp, err := c.GetMySubscriptions(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp, 1)
	assert.Equal(t, 1, resp[0].ID)
	assert.Equal(t, "SUB-123", resp[0].PaypalSubscriptionID)

	mockSubRepo.AssertExpectations(t)
}
