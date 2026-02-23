package controller_test

import (
	"context"
	"testing"
	"time"

	"github.com/demotask/backend/models"
	"github.com/demotask/backend/modules/product"
	"github.com/demotask/backend/modules/product/controller"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/paypal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockProductRepository
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) Create(ctx context.Context, reqData *models.Product, tx *gorm.DB) (*int, error) {
	args := m.Called(ctx, reqData, tx)
	if args.Get(0) != nil {
		id := args.Get(0).(int)
		return &id, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockProductRepository) Find(ctx context.Context, reqData *models.Product) (*models.Product, error) {
	args := m.Called(ctx, reqData)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Product), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockProductRepository) CreateOneTimePrice(ctx context.Context, reqData *models.ProductOneTimePrice, tx *gorm.DB) (*int, error) {
	args := m.Called(ctx, reqData, tx)
	if args.Get(0) != nil {
		id := args.Get(0).(int)
		return &id, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockProductRepository) FindAll(ctx context.Context, filter *models.Product) ([]models.Product, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]models.Product), args.Error(1)
}

func (m *MockProductRepository) FindByIDs(ctx context.Context, ids []int) ([]models.Product, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]models.Product), args.Error(1)
}

func (m *MockProductRepository) UpdatePlanID(ctx context.Context, productID int64, planID string, tx *gorm.DB) error {
	args := m.Called(ctx, productID, planID, tx)
	return args.Error(0)
}

func (m *MockProductRepository) FindAllWithPrice(ctx context.Context) ([]models.Product, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Product), args.Error(1)
}

// MockPayPalClient
type MockPayPalClient struct {
	mock.Mock
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
	return nil, nil
}
func (m *MockPayPalClient) CreateProduct(req *paypal.CreateProductRequest) (*paypal.CreateProductResponse, error) {
	args := m.Called(req)
	if args.Get(0) != nil {
		return args.Get(0).(*paypal.CreateProductResponse), args.Error(1)
	}
	return nil, args.Error(1)
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

func TestCreateProduct(t *testing.T) {
	mockRepo := new(MockProductRepository)
	mockPayPal := new(MockPayPalClient)
	log := logger.NewLogger()

	c := controller.NewController(controller.ProductController{
		Logger:       log,
		Repository:   mockRepo,
		PayPalClient: mockPayPal,
	})

	ctx := context.Background()
	req := &product.CreateProductRequest{
		Name:         "AI Agent",
		Description:  "Smart Agent",
		BusinessType: "DIGITAL",
		Category:     "SOFTWARE",
		ImageURL:     "url",
		HomeURL:      "url",
	}

	mockPayPal.On("CreateProduct", mock.MatchedBy(func(pReq *paypal.CreateProductRequest) bool {
		return pReq.Name == "AI Agent" && pReq.Type == "DIGITAL"
	})).Return(&paypal.CreateProductResponse{ID: "PROD-123"}, nil)

	mockRepo.On("Create", ctx, mock.MatchedBy(func(p *models.Product) bool {
		return p.PaypalProductID == "PROD-123" && p.Name == "AI Agent"
	}), (*gorm.DB)(nil)).Return(1, nil)

	resp, err := c.CreateProduct(ctx, req, nil)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, resp.ID)
	if resp.PaypalProductID != nil {
		assert.Equal(t, "PROD-123", *resp.PaypalProductID)
	}

	mockPayPal.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestGetProducts(t *testing.T) {
	mockRepo := new(MockProductRepository)
	mockPayPal := new(MockPayPalClient)
	log := logger.NewLogger()

	c := controller.NewController(controller.ProductController{
		Logger:       log,
		Repository:   mockRepo,
		PayPalClient: mockPayPal,
	})

	ctx := context.Background()
	now := time.Now()

	products := []models.Product{
		{
			ID:           1,
			Name:         "AI Agent",
			BusinessType: "DIGITAL",
			OneTimePrices: []models.ProductOneTimePrice{
				{Price: 10, Currency: "USD"},
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	mockRepo.On("FindAll", ctx, (*models.Product)(nil)).Return(products, nil)

	req := product.GetProductRequest(&models.Product{})
	resp, err := c.GetProducts(ctx, &req)

	assert.NoError(t, err)
	assert.Len(t, resp, 1)
	assert.Equal(t, 1, resp[0].Product.ID)
	assert.Equal(t, "AI Agent", resp[0].Product.Name)
	assert.NotNil(t, resp[0].OneTimePrice)
	assert.Equal(t, float64(10), resp[0].OneTimePrice.Price)
	assert.Equal(t, "USD", resp[0].OneTimePrice.Currency)

	mockRepo.AssertExpectations(t)
}
