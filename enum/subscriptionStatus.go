package enum

import (
	"fmt"
	"strings"

	"github.com/demotask/backend/utilities"
)

type SubscriptionStatus string

const (
	SubscriptionStatusApprovalPending SubscriptionStatus = "APPROVAL_PENDING"
	SubscriptionStatusActive          SubscriptionStatus = "ACTIVE"
	SubscriptionStatusCancelled       SubscriptionStatus = "CANCELLED"
	SubscriptionStatusSuspended       SubscriptionStatus = "SUSPENDED"
)

func (t SubscriptionStatus) String() string {
	return string(t)
}

func (t SubscriptionStatus) IsValid() error {
	switch t {
	case SubscriptionStatusActive, SubscriptionStatusCancelled, SubscriptionStatusSuspended, SubscriptionStatusApprovalPending:
		return nil
	}
	return fmt.Errorf(utilities.DataNotFound, "Subscription Status")
}

func SubscriptionStatusTypeToArray(value string) ([]string, error) {
	var result []string
	if value == "" {
		return result, nil
	}

	statusSubscriptionType := strings.Split(value, ",")
	for _, t := range statusSubscriptionType {
		trimmed := strings.ToUpper(strings.TrimSpace(t))
		itemType := SubscriptionStatus(trimmed)
		if err := itemType.IsValid(); err != nil {
			return nil, err
		}
		result = append(result, string(itemType))
	}
	return result, nil
}
