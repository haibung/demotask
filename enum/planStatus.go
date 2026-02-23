package enum

import (
	"fmt"
	"strings"

	"github.com/demotask/backend/utilities"
)

type PlanStatus string

const (
	PlanStatusActive   PlanStatus = "ACTIVE"
	PlanStatusInactive PlanStatus = "INACTIVE"
)

func (t PlanStatus) String() string {
	return string(t)
}

func (t PlanStatus) IsValid() error {
	switch t {
	case PlanStatusActive, PlanStatusInactive:
		return nil
	}
	return fmt.Errorf(utilities.DataNotFound, "Plan Status")
}

func PlanStatusTypeToArray(value string) ([]string, error) {
	var result []string
	if value == "" {
		return result, nil
	}

	statusPlanType := strings.Split(value, ",")
	for _, t := range statusPlanType {
		trimmed := strings.ToUpper(strings.TrimSpace(t))
		itemType := PlanStatus(trimmed)
		if err := itemType.IsValid(); err != nil {
			return nil, err
		}
		result = append(result, string(itemType))
	}
	return result, nil
}
