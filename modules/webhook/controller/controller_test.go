package controller_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/demotask/backend/models"
	"github.com/demotask/backend/modules/webhook/controller"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/paypal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockWebhookRepository
type MockWebhookRepository struct{ mock.Mock }

func (m *MockWebhookRepository) FindByEventID(ctx context.Context, eventID string) (*models.WebhookEvent, error) {
	args := m.Called(ctx, eventID)
	if args.Get(0) != nil {
		return args.Get(0).(*models.WebhookEvent), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockWebhookRepository) Save(ctx context.Context, event *models.WebhookEvent, tx *gorm.DB) (*int, error) {
	args := m.Called(ctx, event, tx)
	if args.Get(0) != nil {
		id := args.Get(0).(int)
		return &id, args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockWebhookRepository) MarkProcessed(ctx context.Context, id int, tx *gorm.DB) error {
	args := m.Called(ctx, id, tx)
	return args.Error(0)
}
func (m *MockWebhookRepository) MarkFailed(ctx context.Context, id int, tx *gorm.DB) error {
	args := m.Called(ctx, id, tx)
	return args.Error(0)
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
	return nil, nil
}
func (m *MockPayPalClient) VerifyWebhookSignature(req *paypal.WebhookVerificationRequest) (bool, error) {
	args := m.Called(req)
	return args.Bool(0), args.Error(1)
}
func (m *MockPayPalClient) GetWebhookID() string {
	args := m.Called()
	return args.String(0)
}
func (m *MockPayPalClient) DeactivatePlan(planID string) error { return nil }
func (m *MockPayPalClient) GetSubscriptionDetail(subscriptionID string) (*paypal.SubscriptionDetailResponse, error) {
	return nil, nil
}

// MockSubscriptionHandler
type MockSubscriptionHandler struct{ mock.Mock }

func (m *MockSubscriptionHandler) Handle(ctx context.Context, eventType string, body map[string]interface{}, tx *gorm.DB) error {
	args := m.Called(ctx, eventType, body, tx)
	return args.Error(0)
}
func (m *MockSubscriptionHandler) HandleSale(ctx context.Context, body map[string]interface{}, tx *gorm.DB) error {
	args := m.Called(ctx, body, tx)
	return args.Error(0)
}
func (m *MockSubscriptionHandler) HandleSubscription(ctx context.Context, eventType string, body map[string]interface{}, tx *gorm.DB) error {
	args := m.Called(ctx, eventType, body, tx)
	return args.Error(0)
}

// MockPaymentCaptureHandler
type MockPaymentCaptureHandler struct{ mock.Mock }

func (m *MockPaymentCaptureHandler) Handle(ctx context.Context, body map[string]interface{}, tx *gorm.DB) error {
	args := m.Called(ctx, body, tx)
	return args.Error(0)
}
func (m *MockPaymentCaptureHandler) HandleVault(ctx context.Context, body map[string]interface{}, tx *gorm.DB) error {
	args := m.Called(ctx, body, tx)
	return args.Error(0)
}
func (m *MockPaymentCaptureHandler) HandleOrderApproved(ctx context.Context, event paypal.WebhookEvent, tx *gorm.DB) error {
	args := m.Called(ctx, event, tx)
	return args.Error(0)
}
func (m *MockPaymentCaptureHandler) HandleCapture(ctx context.Context, body map[string]interface{}, tx *gorm.DB) error {
	args := m.Called(ctx, body, tx)
	return args.Error(0)
}

func TestHandlePayPalWebhook(t *testing.T) {
	mockWebhookRepo := new(MockWebhookRepository)
	mockPayPal := new(MockPayPalClient)
	mockSubHandler := new(MockSubscriptionHandler)
	mockPayHandler := new(MockPaymentCaptureHandler)
	log := logger.NewLogger()

	c := controller.NewController(controller.WebhookController{
		Logger:              log,
		WebhookRepo:         mockWebhookRepo,
		PayPalClient:        mockPayPal,
		SubscriptionHandler: mockSubHandler,
		PaymentHandler:      mockPayHandler,
	})

	ctx := context.Background()

	eventID := "EVT-123"
	eventType := "PAYMENT.CAPTURE.COMPLETED"

	body := map[string]interface{}{
		"id":         eventID,
		"event_type": eventType,
	}
	rawBody, _ := json.Marshal(body)

	headers := map[string]string{
		"paypal-auth-algo":         "alg",
		"paypal-cert-url":          "url",
		"paypal-transmission-id":   "tid",
		"paypal-transmission-sig":  "sig",
		"paypal-transmission-time": "time",
	}

	mockWebhookRepo.On("FindByEventID", ctx, eventID).Return((*models.WebhookEvent)(nil), gorm.ErrRecordNotFound)
	mockPayPal.On("GetWebhookID").Return("WH-123")

	verifyReq := &paypal.WebhookVerificationRequest{
		AuthAlgo:         "alg",
		CertURL:          "url",
		TransmissionID:   "tid",
		TransmissionSig:  "sig",
		TransmissionTime: "time",
		WebhookID:        "WH-123",
		WebhookEvent:     rawBody,
	}
	mockPayPal.On("VerifyWebhookSignature", verifyReq).Return(true, nil)

	mockWebhookRepo.On("Save", ctx, mock.AnythingOfType("*models.WebhookEvent"), (*gorm.DB)(nil)).Return(1, nil)
	mockPayHandler.On("HandleCapture", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockWebhookRepo.On("MarkProcessed", ctx, 1, (*gorm.DB)(nil)).Return(nil)

	err := c.HandlePayPalWebhook(ctx, headers, rawBody, body, nil)

	assert.NoError(t, err)

	mockWebhookRepo.AssertExpectations(t)
	mockPayPal.AssertExpectations(t)
	mockPayHandler.AssertExpectations(t)
}
