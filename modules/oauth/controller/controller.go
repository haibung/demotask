package controller

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/demotask/backend/modules/oauth"
	"github.com/demotask/backend/modules/oauth/repository"

	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/paypal"
	"github.com/demotask/backend/utilities"
	"go.uber.org/fx"
)

type (
	IOAuthController interface {
		GetClientToken(ctx context.Context) (*oauth.TokenResponse, error)
	}

	OAuthController struct {
		fx.In
		Logger          *logger.Logger
		OAuthRepository repository.OAuthRepository
		PayPalClient    paypal.IPayPalClient
	}
)

// NewController creates a new OAuth controller
func NewController(oauthController OAuthController) IOAuthController {
	return &oauthController
}

// GetClientToken retrieves a PayPal access token using client credentials
func (receiver *OAuthController) GetClientToken(ctx context.Context) (*oauth.TokenResponse, error) {
	// Use PayPal client to get token using client credentials
	tokenResp, err := receiver.PayPalClient.ClientCredentials()
	if err != nil {
		receiver.Logger.Error(err)
		return nil, utilities.ErrorRequest(
			fmt.Errorf("failed to get client credentials token: %w", err),
			http.StatusInternalServerError,
		)
	}

	return &oauth.TokenResponse{
		AccessToken: tokenResp.AccessToken,
		TokenType:   tokenResp.TokenType,
		AppID:       tokenResp.AppID,
		ExpiresIn:   tokenResp.ExpiresIn,
		Nonce:       tokenResp.Nonce,
		Scope:       tokenResp.Scope,
	}, nil
}

// generateStateToken generates a random state token for CSRF protection
func generateStateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// // InitiateOAuth generates the OAuth authorization URL
// func (receiver *OAuthController) InitiateOAuth(ctx context.Context, reqData *oauth.OAuthInitiateRequest) (*oauth.OAuthInitiateResponse, error) {
// 	// Validate request data
// 	messages, err := reqData.Validate()
// 	if err != nil {
// 		receiver.Logger.Error(err)
// 		return nil, utilities.ErrorRequest(err, http.StatusBadRequest, messages)
// 	}

// 	// Generate CSRF state token
// 	state, err := generateStateToken()
// 	if err != nil {
// 		receiver.Logger.Error(err)
// 		return nil, utilities.ErrorRequest(err, http.StatusInternalServerError)
// 	}

// 	var authURL string
// 	switch reqData.Provider {
// 	case "paypal":
// 		authURL = receiver.PayPalClient.GetAuthorizationURL(state)
// 	default:
// 		return nil, utilities.ErrorRequest(
// 			fmt.Errorf("unsupported OAuth provider: %s", reqData.Provider),
// 			http.StatusBadRequest,
// 		)
// 	}

// 	return &oauth.OAuthInitiateResponse{
// 		AuthorizationURL: authURL,
// 		State:            state,
// 	}, nil
// }

// func (receiver *OAuthController) HandleCallback(ctx context.Context, reqData *oauth.OAuthCallbackRequest, tx *gorm.DB) (*oauth.OAuthCallbackResponse, error) {
// 	messages, err := reqData.Validate()
// 	if err != nil {
// 		receiver.Logger.Error(err)
// 		return nil, utilities.ErrorRequest(err, http.StatusBadRequest, messages)
// 	}

// 	var tokenResp *paypal.TokenResponse

// 	switch reqData.Provider {
// 	case "paypal":
// 		tokenResp, err = receiver.PayPalClient.ExchangeCodeForToken(reqData.Code)
// 		if err != nil {
// 			receiver.Logger.Error(err)
// 			return nil, utilities.ErrorRequest(
// 				fmt.Errorf("failed to exchange code for token: %w", err),
// 				http.StatusInternalServerError,
// 			)
// 		}
// 	default:
// 		return nil, utilities.ErrorRequest(
// 			fmt.Errorf("unsupported OAuth provider: %s", reqData.Provider),
// 			http.StatusBadRequest,
// 		)
// 	}

// 	// Return PayPal tokens (users are created from webhooks, not OAuth callback)
// 	return &oauth.OAuthCallbackResponse{
// 		AccessToken:  tokenResp.AccessToken,
// 		RefreshToken: tokenResp.RefreshToken,
// 		TokenType:    "Bearer",
// 		ExpiresIn:    tokenResp.ExpiresIn,
// 		Provider:     "paypal",
// 	}, nil
// }
