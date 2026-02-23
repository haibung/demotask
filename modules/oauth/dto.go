package oauth

import (
	"errors"

	"github.com/demotask/backend/packages/validation"
	"github.com/demotask/backend/utilities"
)

type (
	// Struct Request
	OAuthInitiateRequest struct {
		ContextUserID int

		Provider string
		State    string

		Validation validation.Validation
	}

	OAuthCallbackRequest struct {
		ContextUserID int

		Provider string
		Code     string
		State    string

		Validation validation.Validation
	}

	DirectOAuthRequest struct {
		ContextUserID int

		Provider    string
		AccessToken string `json:"access_token"`

		Validation validation.Validation
	}

	LinkAccountRequest struct {
		ContextUserID int

		Provider string
		Code     string

		Validation validation.Validation
	}

	TokenRefreshRequest struct {
		ContextUserID int

		Provider     string
		RefreshToken string

		Validation validation.Validation
	}

	UnlinkAccountRequest struct {
		ContextUserID int

		Provider string

		Validation validation.Validation
	}

	// Struct Response
	OAuthInitiateResponse struct {
		AuthorizationURL string `json:"authorization_url"`
		State            string `json:"state"`
	}

	OAuthCallbackResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		UserID       int    `json:"user_id"`
		IsNewUser    bool   `json:"is_new_user"`
		Provider     string `json:"provider"`
	}

	TokenRefreshResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	TokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		AppID       string `json:"app_id"`
		ExpiresIn   int    `json:"expires_in"`
		Nonce       string `json:"nonce"`
		Scope       string `json:"scope"`
	}
)

func (receiver OAuthInitiateRequest) Validate() ([]map[string]interface{}, error) {
	receiver.Validation.IsEmptyString(receiver.Provider, "Provider")

	if len(receiver.Validation.Messages) > 0 {
		return receiver.Validation.Messages, errors.New(utilities.BadRequest)
	}

	return nil, nil
}

func (receiver OAuthCallbackRequest) Validate() ([]map[string]interface{}, error) {
	receiver.Validation.IsEmptyString(receiver.Provider, "Provider")
	receiver.Validation.IsEmptyString(receiver.Code, "Code")
	receiver.Validation.IsEmptyString(receiver.State, "State")

	if len(receiver.Validation.Messages) > 0 {
		return receiver.Validation.Messages, errors.New(utilities.BadRequest)
	}

	return nil, nil
}

func (receiver DirectOAuthRequest) Validate() ([]map[string]interface{}, error) {
	receiver.Validation.IsEmptyString(receiver.Provider, "Provider")
	receiver.Validation.IsEmptyString(receiver.AccessToken, "AccessToken")

	if len(receiver.Validation.Messages) > 0 {
		return receiver.Validation.Messages, errors.New(utilities.BadRequest)
	}

	return nil, nil
}

func (receiver LinkAccountRequest) Validate() ([]map[string]interface{}, error) {
	receiver.Validation.IsEmptyString(receiver.Provider, "Provider")
	receiver.Validation.IsEmptyString(receiver.Code, "Code")

	if len(receiver.Validation.Messages) > 0 {
		return receiver.Validation.Messages, errors.New(utilities.BadRequest)
	}

	return nil, nil
}

func (receiver TokenRefreshRequest) Validate() ([]map[string]interface{}, error) {
	receiver.Validation.IsEmptyString(receiver.Provider, "Provider")
	receiver.Validation.IsEmptyString(receiver.RefreshToken, "RefreshToken")

	if len(receiver.Validation.Messages) > 0 {
		return receiver.Validation.Messages, errors.New(utilities.BadRequest)
	}

	return nil, nil
}

func (receiver UnlinkAccountRequest) Validate() ([]map[string]interface{}, error) {
	receiver.Validation.IsEmptyString(receiver.Provider, "Provider")

	if len(receiver.Validation.Messages) > 0 {
		return receiver.Validation.Messages, errors.New(utilities.BadRequest)
	}

	return nil, nil
}
