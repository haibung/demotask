package controller_test

import (
	"context"
	"testing"

	"github.com/demotask/backend/modules/oauth/controller"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/paypal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPayPalClient
type MockPayPalClient struct{ mock.Mock }

func (m *MockPayPalClient) GetAuthorizationURL(state string) string { return "" }
func (m *MockPayPalClient) ExchangeCodeForToken(code string) (*paypal.TokenResponse, error) {
	return nil, nil
}
func (m *MockPayPalClient) RefreshAccessToken(refreshToken string) (*paypal.TokenResponse, error) {
	return nil, nil
}
func (m *MockPayPalClient) ClientCredentials() (*paypal.TokenResponse, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).(*paypal.TokenResponse), args.Error(1)
	}
	return nil, args.Error(1)
}
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
	return false, nil
}
func (m *MockPayPalClient) GetWebhookID() string               { return "" }
func (m *MockPayPalClient) DeactivatePlan(planID string) error { return nil }
func (m *MockPayPalClient) GetSubscriptionDetail(subscriptionID string) (*paypal.SubscriptionDetailResponse, error) {
	return nil, nil
}

func TestGetClientToken(t *testing.T) {
	mockPayPal := new(MockPayPalClient)
	log := logger.NewLogger()

	c := controller.NewController(controller.OAuthController{
		PayPalClient: mockPayPal,
		Logger:       log,
	})

	ctx := context.Background()

	mockResp := &paypal.TokenResponse{
		AccessToken: "test-access-token",
		TokenType:   "Bearer",
		AppID:       "APP-123",
		ExpiresIn:   3600,
	}

	mockPayPal.On("ClientCredentials").Return(mockResp, nil)

	resp, err := c.GetClientToken(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test-access-token", resp.AccessToken)
	assert.Equal(t, "Bearer", resp.TokenType)
	assert.Equal(t, "APP-123", resp.AppID)
	assert.Equal(t, 3600, resp.ExpiresIn)

	mockPayPal.AssertExpectations(t)
}
