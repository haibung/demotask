package user

import (
	"errors"
	"time"

	"github.com/demotask/backend/packages/validation"
	"github.com/demotask/backend/utilities"
)

type (
	//Struct Request
	CreateRequest struct {
		ContextUserID int

		FullName string
		Password string
		Email    string

		Validation validation.Validation
	}

	FindByEmailRequest struct {
		ContextUserID int

		Email string

		Validation validation.Validation
	}

	//	Struct Response
	FindByEmailResponse struct {
		ID        int       `json:"id"`
		FullName  string    `json:"full_name"`
		Email     string    `json:"email"`
		Password  string    `json:"password"`
		CreatedAt time.Time `json:"created_at"`
	}
)

func (receiver CreateRequest) Validate() ([]map[string]interface{}, error) {
	receiver.Validation.IsEmptyString(receiver.FullName, "FullName")
	receiver.Validation.IsEmptyString(receiver.Password, "Password")
	receiver.Validation.IsEmptyString(receiver.Email, "Email")
	receiver.Validation.IsEmailValid(receiver.Email, "Email")

	if len(receiver.Validation.Messages) > 0 {
		return receiver.Validation.Messages, errors.New(utilities.BadRequest)
	}

	return nil, nil
}

func (receiver FindByEmailRequest) Validate() ([]map[string]interface{}, error) {
	receiver.Validation.IsEmptyString(receiver.Email, "Email")
	receiver.Validation.IsEmailValid(receiver.Email, "Email")

	if len(receiver.Validation.Messages) > 0 {
		return receiver.Validation.Messages, errors.New(utilities.BadRequest)
	}

	return nil, nil
}
