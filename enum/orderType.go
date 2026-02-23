package enum

import (
	"fmt"
	"strings"

	"github.com/demotask/backend/utilities"
)

type OrderType string

const (
	OrderTypeSubscription OrderType = "SUBSCRIPTION"
	OrderTypeOneTime      OrderType = "ONE_TIME"
)

func (t OrderType) String() string {
	return string(t)
}

func (t OrderType) IsValid() error {
	switch t {
	case OrderTypeSubscription, OrderTypeOneTime:
		return nil
	}
	return fmt.Errorf(utilities.DataNotFound, "Order Type")
}

func OrderTypeToArrayString(value string) ([]string, error) {
	var result []string
	if value == "" {
		return result, nil
	}

	statusPlanType := strings.Split(value, ",")
	for _, t := range statusPlanType {
		trimmed := strings.ToUpper(strings.TrimSpace(t))
		itemType := OrderType(trimmed)
		if err := itemType.IsValid(); err != nil {
			return nil, err
		}
		result = append(result, string(itemType))
	}
	return result, nil
}
