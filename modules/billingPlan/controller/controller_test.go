package controller_test

import (
	"context"
	"testing"

	"github.com/demotask/backend/enum"
	"github.com/demotask/backend/models"
	"github.com/demotask/backend/modules/billingPlan"
	"github.com/demotask/backend/modules/billingPlan/controller"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/paypal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockBillingPlanRepository
type MockBillingPlanRepository struct{ mock.Mock }

func (m *MockBillingPlanRepository) ExistsActivePlan(ctx context.Context, productID int, name string) (bool, error) {
	return false, nil
}
func (m *MockBillingPlanRepository) Create(ctx context.Context, req *models.BillingPlan, tx *gorm.DB) (*int, error) {
	args := m.Called(ctx, req, tx)
	if args.Get(0) != nil {
		id := args.Get(0).(int)
		return &id, args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockBillingPlanRepository) CreateBillingCycle(ctx context.Context, req *models.BillingCycle, tx *gorm.DB) error {
	args := m.Called(ctx, req, tx)
	return args.Error(0)
}
func (m *MockBillingPlanRepository) FindByID(ctx context.Context, req *models.BillingPlan) (*models.BillingPlan, error) {
	args := m.Called(ctx, req)
	if args.Get(0) != nil {
		return args.Get(0).(*models.BillingPlan), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockBillingPlanRepository) FindByPayPalID(ctx context.Context, paypalSubscriptionID string) (*models.Subscription, error) {
	args := m.Called(ctx, paypalSubscriptionID)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Subscription), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockBillingPlanRepository) UpdateStatus(ctx context.Context, paypalSubscriptionID string, status string, tx *gorm.DB) error {
	args := m.Called(ctx, paypalSubscriptionID, status, tx)
	return args.Error(0)
}

// MockProductRepository
type MockProductRepository struct{ mock.Mock }

func (m *MockProductRepository) Create(ctx context.Context, req *models.Product, tx *gorm.DB) (*int, error) {
	return nil, nil
}
func (m *MockProductRepository) Find(ctx context.Context, req *models.Product) (*models.Product, error) {
	args := m.Called(ctx, req)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Product), args.Error(1)
	}
	return nil, args.Error(1)
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
	args := m.Called(req, requestID)
	if args.Get(0) != nil {
		return args.Get(0).(*paypal.CreatePlanResponse), args.Error(1)
	}
	return nil, args.Error(1)
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

func TestCreateBillingPlan(t *testing.T) {
	mockBillingRepo := new(MockBillingPlanRepository)
	mockProductRepo := new(MockProductRepository)
	mockPayPal := new(MockPayPalClient)
	log := logger.NewLogger()

	c := controller.NewController(controller.BillingPlanController{
		Logger:                log,
		BillingPlanRepository: mockBillingRepo,
		ProductRepository:     mockProductRepo,
		PayPalClient:          mockPayPal,
	})

	ctx := context.Background()
	productID := 1

	req := &billingPlan.CreateBillingPlanRequest{
		ProductID:   &productID,
		Name:        "Test Plan",
		Description: "A test plan",
		Status:      enum.PlanStatusActive,
		BillingCycles: []billingPlan.CreateBillingCycleRequest{
			{
				IntervalUnit:  "MONTH",
				IntervalCount: 1,
				TenureType:    "REGULAR",
				Sequence:      1,
				TotalCycles:   0,
				Price:         10.0,
				Currency:      "USD",
			},
		},
	}

	mockProductRepo.On("Find", ctx, mock.MatchedBy(func(p *models.Product) bool {
		return p.ID == productID
	})).Return(&models.Product{ID: productID, PaypalProductID: "PROD-123"}, nil)

	mockPayPal.On("CreatePlan", mock.AnythingOfType("*paypal.CreatePlanRequest"), mock.AnythingOfType("string")).Return(&paypal.CreatePlanResponse{ID: "PLAN-123"}, nil)

	mockBillingRepo.On("Create", ctx, mock.AnythingOfType("*models.BillingPlan"), (*gorm.DB)(nil)).Return(1, nil).Run(func(args mock.Arguments) {
		plan := args.Get(1).(*models.BillingPlan)
		plan.ID = 1
	})
	mockBillingRepo.On("CreateBillingCycle", ctx, mock.AnythingOfType("*models.BillingCycle"), (*gorm.DB)(nil)).Return(nil)

	resp, err := c.CreateBillingPlan(ctx, req, nil)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, resp.ID)
	assert.Equal(t, "PLAN-123", resp.PaypalPlanID)

	mockBillingRepo.AssertExpectations(t)
	mockProductRepo.AssertExpectations(t)
	mockPayPal.AssertExpectations(t)
}
