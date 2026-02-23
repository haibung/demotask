package auth

import (
	"errors"
	"fmt"

	"github.com/demotask/backend/packages/validation"
	"github.com/demotask/backend/utilities"
)

type (
	// Struct Request
	LoginRequest struct {
		ContextUserID int

		Email    string
		Password string

		Validation validation.Validation
	}

	//	Struct Response
	LoginResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
)

type RegisterRequest struct {
	FullName   string
	Email      string
	Password   string
	Validation validation.Validation
}

func (receiver RegisterRequest) Validate() ([]map[string]interface{}, error) {
	receiver.Validation.IsEmptyString(receiver.Password, "Password")
	receiver.Validation.IsEmptyString(receiver.Email, "Email")
	receiver.Validation.IsEmptyString(receiver.FullName, "Full Name")
	if !utilities.EmailRegex.MatchString(receiver.Email) {
		return nil, fmt.Errorf(utilities.ValueNotValid, "Email")
	}

	if len(receiver.Validation.Messages) > 0 {
		return receiver.Validation.Messages, errors.New(utilities.BadRequest)
	}

	return nil, nil
}

func (receiver LoginRequest) Validate() ([]map[string]interface{}, error) {
	receiver.Validation.IsEmptyString(receiver.Password, "Password")
	receiver.Validation.IsEmptyString(receiver.Email, "Email")

	if len(receiver.Validation.Messages) > 0 {
		return receiver.Validation.Messages, errors.New(utilities.BadRequest)
	}

	return nil, nil
}
