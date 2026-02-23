package enum

import (
	"fmt"
	"strings"

	"github.com/demotask/backend/utilities"
)

type WebhookStatus string

const (
	WebhookStatusPending   WebhookStatus = "PENDING"
	WebhookStatusProcessed WebhookStatus = "PROCESSED"
	WebhookStatusFailed    WebhookStatus = "FAILED"
)

func (t WebhookStatus) String() string {
	return string(t)
}

func (t WebhookStatus) IsValid() error {
	switch t {
	case WebhookStatusPending, WebhookStatusProcessed, WebhookStatusFailed:
		return nil
	}
	return fmt.Errorf(utilities.DataNotFound, "Webhook Status")
}

func WebhookStatusTypeToArray(value string) ([]string, error) {
	var result []string
	if value == "" {
		return result, nil
	}

	statusSubscriptionType := strings.Split(value, ",")
	for _, t := range statusSubscriptionType {
		trimmed := strings.ToUpper(strings.TrimSpace(t))
		itemType := WebhookStatus(trimmed)
		if err := itemType.IsValid(); err != nil {
			return nil, err
		}
		result = append(result, string(itemType))
	}
	return result, nil
}
