package paypal

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/demotask/backend/config"
	"github.com/demotask/backend/packages/logger"
	"go.uber.org/fx"
)

// PayPal API endpoints
const (
	endpointToken              = "/v1/oauth2/token"
	endpointCreateOrder        = "/v2/checkout/orders"
	endpointCaptureOrder       = "/v2/checkout/orders/%s/capture"
	endpointVerifyWebhook      = "/v1/notifications/verify-webhook-signature"
	endpointCreateSubscription = "/v1/billing/subscriptions"
	endpointCreateProduct      = "/v1/catalogs/products"
	endpointCreatePlan         = "/v1/billing/plans"
	endpointVaultSetup         = "/v3/vault/setup-tokens"
	endpointVaultPaymentToken  = "/v3/vault/payment-tokens"
	// endpointDeactivatePlan     = "/v1/billing/plans/%s/deactivate"
)

type (
	IPayPalClient interface {
		// OAuth methods
		GetAuthorizationURL(state string) string
		ExchangeCodeForToken(code string) (*TokenResponse, error)
		RefreshAccessToken(refreshToken string) (*TokenResponse, error)
		ClientCredentials() (*TokenResponse, error)
		GetUserInfo(accessToken string) (*UserInfo, error)
		RevokeToken(token string) error

		// Plan methods
		DeactivatePlan(planID string) error

		// Payment methods
		CreateOrder(req *OrderRequest) (*OrderResponse, error)
		CaptureOrder(orderID string) (*CaptureResponse, error)

		// Subscription methods
		CreateSubscription(req *CreateSubscriptionRequest) (*CreateSubscriptionResponse, error)
		CreateProduct(req *CreateProductRequest) (*CreateProductResponse, error)
		CreatePlan(req *CreatePlanRequest, requestID string) (*CreatePlanResponse, error)
		GetSubscriptionDetail(subscriptionID string) (*SubscriptionDetailResponse, error)

		// Vaulting methods
		CreateVaultSetupToken(req *VaultSetupTokenRequest) (*VaultSetupTokenResponse, error)
		CreatePaymentToken(req *CreatePaymentTokenRequest) (*CreatePaymentTokenResponse, error)
		CreateOrderWithVault(amount float64, currency string, vaultID string) (*CreateOrderResponse, error)

		// Webhook methods
		VerifyWebhookSignature(req *WebhookVerificationRequest) (bool, error)
		GetWebhookID() string
	}

	PayPalClientParams struct {
		fx.In
		Logger *logger.Logger
		Config *config.Config
	}

	PayPalClient struct {
		Logger       *logger.Logger
		BaseURL      string
		AuthBaseURL  string
		ClientID     string
		ClientSecret string
		RedirectURL  string
		WebhookID    string
		httpClient   *http.Client
		accessToken  string
		expiry       time.Time
		mutex        sync.Mutex
	}
)

func NewPayPalClient(p PayPalClientParams) IPayPalClient {
	cfg := p.Config.PayPal

	return &PayPalClient{
		Logger:       p.Logger,
		BaseURL:      cfg.BaseURL,
		AuthBaseURL:  cfg.AuthBaseURL,
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		WebhookID:    cfg.WebhookID,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *PayPalClient) GetAuthorizationURL(state string) string {
	params := url.Values{}
	params.Set("client_id", c.ClientID)
	params.Set("response_type", "code")
	params.Set("scope", "openid profile email")
	params.Set("redirect_uri", c.RedirectURL)
	params.Set("state", state)

	return fmt.Sprintf(
		"%s/signin/authorize?%s",
		c.AuthBaseURL,
		params.Encode(),
	)
}

func (c *PayPalClient) GetWebhookID() string {
	return c.WebhookID
}

func (c *PayPalClient) getAccessToken() (string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.accessToken != "" && time.Now().Before(c.expiry.Add(-1*time.Minute)) {
		return c.accessToken, nil
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest(
		http.MethodPost,
		c.BaseURL+endpointToken,
		bytes.NewBufferString(data.Encode()),
	)
	if err != nil {
		return "", err
	}

	auth := base64.StdEncoding.EncodeToString(
		[]byte(c.ClientID + ":" + c.ClientSecret),
	)

	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("paypal token error %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}

	c.accessToken = tokenResp.AccessToken
	c.expiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return c.accessToken, nil
}

func (c *PayPalClient) doRequest(method string, endpoint string, body interface{}, auth bool, result interface{}, headers map[string]string) error {
	var reader io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, c.BaseURL+endpoint, reader)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if auth {
		token, err := c.getAccessToken()
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 300 {
		return fmt.Errorf("paypal error %d: %s", resp.StatusCode, string(respBytes))
	}

	if result != nil {
		if err := json.Unmarshal(respBytes, result); err != nil {
			return err
		}
	}

	return nil
}

func (c *PayPalClient) CreateOrder(req *OrderRequest) (*OrderResponse, error) {
	var resp OrderResponse

	err := c.doRequest(
		http.MethodPost,
		endpointCreateOrder,
		req,
		true,
		&resp,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *PayPalClient) CaptureOrder(orderID string) (*CaptureResponse, error) {
	var resp CaptureResponse

	err := c.doRequest(
		http.MethodPost,
		fmt.Sprintf(endpointCaptureOrder, orderID),
		nil,
		true,
		&resp,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *PayPalClient) VerifyWebhookSignature(req *WebhookVerificationRequest) (bool, error) {
	var resp WebhookVerificationResponse
	err := c.doRequest(
		http.MethodPost,
		endpointVerifyWebhook,
		req,
		true,
		&resp,
		nil,
	)
	if err != nil {
		return false, err
	}

	return resp.VerificationStatus == "SUCCESS", nil
}

func (c *PayPalClient) CreateSubscription(req *CreateSubscriptionRequest) (*CreateSubscriptionResponse, error) {
	var resp CreateSubscriptionResponse
	err := c.doRequest(
		http.MethodPost,
		endpointCreateSubscription,
		req,
		true,
		&resp,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *PayPalClient) CreateProduct(req *CreateProductRequest) (*CreateProductResponse, error) {
	var resp CreateProductResponse
	headers := map[string]string{
		"PayPal-Request-Id": fmt.Sprintf("PRODUCT-%d", time.Now().UnixNano()),
	}

	err := c.doRequest(
		http.MethodPost,
		endpointCreateProduct,
		req,
		true,
		&resp,
		headers,
	)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *PayPalClient) CreatePlan(req *CreatePlanRequest, requestID string) (*CreatePlanResponse, error) {
	var resp CreatePlanResponse
	headers := map[string]string{
		"Prefer":            "return=representation",
		"PayPal-Request-Id": requestID,
	}

	err := c.doRequest(
		http.MethodPost,
		endpointCreatePlan,
		req,
		true,
		&resp,
		headers,
	)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *PayPalClient) CreateVaultSetupToken(req *VaultSetupTokenRequest) (*VaultSetupTokenResponse, error) {
	var resp VaultSetupTokenResponse
	err := c.doRequest(
		http.MethodPost,
		endpointVaultSetup,
		req,
		true,
		&resp,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *PayPalClient) CreatePaymentToken(req *CreatePaymentTokenRequest) (*CreatePaymentTokenResponse, error) {
	var resp CreatePaymentTokenResponse
	err := c.doRequest(
		http.MethodPost,
		endpointVaultPaymentToken,
		req,
		true,
		&resp,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *PayPalClient) RevokeToken(token string) error {
	revokeURL := fmt.Sprintf("%s/v1/oauth2/token/revoke", c.BaseURL)

	data := url.Values{}
	data.Set("token", token)
	data.Set("token_type_hint", "ACCESS_TOKEN")

	req, err := http.NewRequest("POST", revokeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create revoke request: %w", err)
	}

	auth := base64.StdEncoding.EncodeToString([]byte(c.ClientID + ":" + c.ClientSecret))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("revoke request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("revoke failed: %s", string(body))
	}

	return nil
}

func (c *PayPalClient) requestToken(data url.Values) (*TokenResponse, error) {

	tokenURL := fmt.Sprintf("%s/v1/oauth2/token", c.BaseURL)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	auth := base64.StdEncoding.EncodeToString([]byte(c.ClientID + ":" + c.ClientSecret))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed: %s", string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokenResp, nil
}

func (c *PayPalClient) CreateOrderWithVault(amount float64, currency string, vaultID string) (*CreateOrderResponse, error) {

	req := &VaultOrderRequest{
		Intent: "CAPTURE",
		PurchaseUnits: []VaultPurchaseUnit{
			{
				Amount: VaultAmount{
					CurrencyCode: currency,
					Value:        fmt.Sprintf("%.2f", amount),
				},
			},
		},
		PaymentSource: VaultPaymentSource{
			PayPal: PaypalSource{
				VaultID: vaultID,
			},
		},
	}

	headers := map[string]string{
		"PayPal-Request-Id": fmt.Sprintf("VAULT-%d", time.Now().UnixNano()),
	}

	var resp CreateOrderResponse

	err := c.doRequest(
		http.MethodPost,
		endpointCreateOrder,
		req,
		true,
		&resp,
		headers,
	)

	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *PayPalClient) GetVaultToken(tokenID string) (*VaultTokenDetailsResponse, error) {
	url := fmt.Sprintf("/v3/vault/payment-tokens/%s", tokenID)

	var resp VaultTokenDetailsResponse
	err := c.doRequest(http.MethodGet, url, nil, true, &resp, nil)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ExchangeCodeForToken exchanges authorization code for access and refresh tokens
func (c *PayPalClient) ExchangeCodeForToken(code string) (*TokenResponse, error) {

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", c.RedirectURL)

	return c.requestToken(data)
}

// RefreshAccessToken refreshes an expired access token
func (c *PayPalClient) RefreshAccessToken(refreshToken string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	return c.requestToken(data)
}

func (c *PayPalClient) ClientCredentials() (*TokenResponse, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.accessToken != "" && time.Now().Before(c.expiry.Add(-1*time.Minute)) {
		return &TokenResponse{
			AccessToken: c.accessToken,
			ExpiresIn:   int(time.Until(c.expiry).Seconds()),
		}, nil
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	tokenResp, err := c.requestToken(data)
	if err != nil {
		return nil, err
	}
	c.accessToken = tokenResp.AccessToken
	c.expiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return tokenResp, nil
}

func (c *PayPalClient) GetUserInfo(accessToken string) (*UserInfo, error) {
	var resp UserInfo

	headers := map[string]string{
		"Authorization": "Bearer " + accessToken,
	}

	err := c.doRequest(
		http.MethodGet,
		"/v1/identity/openidconnect/userinfo?schema=openid",
		nil,
		false,
		&resp,
		headers,
	)

	if err != nil {
		return nil, err
	}

	return &resp, nil
}
func (c *PayPalClient) DeactivatePlan(planID string) error {

	patchBody := []map[string]interface{}{
		{
			"op":    "replace",
			"path":  "/status",
			"value": "INACTIVE",
		},
	}

	return c.doRequest(
		http.MethodPatch,
		fmt.Sprintf("/v1/billing/plans/%s", planID),
		patchBody,
		true,
		nil,
		nil,
	)
}

func (c *PayPalClient) GetSubscriptionDetail(subscriptionID string) (*SubscriptionDetailResponse, error) {

	var resp SubscriptionDetailResponse

	err := c.doRequest(
		http.MethodGet,
		fmt.Sprintf("/v1/billing/subscriptions/%s", subscriptionID),
		nil,
		true,
		&resp,
		nil,
	)

	if err != nil {
		return nil, err
	}

	return &resp, nil
}
