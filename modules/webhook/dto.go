package webhook

import (
	"errors"

	"github.com/demotask/backend/packages/validation"
	"github.com/demotask/backend/utilities"
)

type (
	// Struct Request
	WebhookPayload struct {
		// ContextUserID int

		Provider string
		Headers  map[string]string
		Body     map[string]interface{}

		Validation validation.Validation
	}

	// Struct Response
	WebhookResponse struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
)

func (receiver WebhookPayload) Validate() ([]map[string]interface{}, error) {
	receiver.Validation.IsEmptyString(receiver.Provider, "Provider")

	if len(receiver.Validation.Messages) > 0 {
		return receiver.Validation.Messages, errors.New(utilities.BadRequest)
	}

	return nil, nil
}
