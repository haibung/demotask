package enum

import (
	"fmt"
	"strings"

	"github.com/demotask/backend/utilities"
)

type BusinessType string

const (
	PayPalTypePhysical BusinessType = "PHYSICAL"
	PayPalTypeDigital  BusinessType = "DIGITAL"
	PayPalTypeService  BusinessType = "SERVICE"
)

func (t BusinessType) String() string {
	return string(t)
}

func (t BusinessType) IsValid() error {
	switch t {
	case PayPalTypePhysical, PayPalTypeDigital, PayPalTypeService:
		return nil
	}
	return fmt.Errorf(utilities.DataNotFound, "PayPal Item Type")
}

func BusinessTypeToString(value string) ([]BusinessType, error) {
	var result []BusinessType
	if value == "" {
		return result, nil
	}

	types := strings.Split(value, ",")
	for _, t := range types {
		trimmed := strings.ToUpper(strings.TrimSpace(t))
		itemType := BusinessType(trimmed)
		if err := itemType.IsValid(); err != nil {
			return nil, err
		}
		result = append(result, itemType)
	}
	return result, nil
}
