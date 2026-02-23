package validation

import (
	"fmt"
	"strings"

	"github.com/demotask/backend/utilities"
)

type Validation struct {
	Messages []map[string]interface{}
}

// IsEmptyString :
func (receiver *Validation) IsEmptyString(value, label string) {
	if value == "" {
		err := fmt.Errorf(utilities.EmptyValue, strings.ToLower(label))
		receiver.Messages = append(receiver.Messages, map[string]interface{}{
			strings.ToLower(label): err.Error(),
		})
	}
}

// IsEmailValid :
func (receiver *Validation) IsEmailValid(value, label string) {
	if !utilities.EmailRegex.MatchString(value) {
		err := fmt.Errorf(utilities.ValueNotValid, strings.ToLower(label))
		receiver.Messages = append(receiver.Messages, map[string]interface{}{
			strings.ToLower(label): err.Error(),
		})
	}
}

// IsIntegerMax :
func (receiver *Validation) IsIntegerMax(value, max int, label string) {
	if value > max {
		err := fmt.Errorf(utilities.MaxValueMust, strings.ToLower(label), max)
		receiver.Messages = append(receiver.Messages, map[string]interface{}{
			strings.ToLower(label): err.Error(),
		})
	}
}

// IsIntegerMin :
func (receiver *Validation) IsIntegerMin(value, min int, label string) {
	if value < min {
		err := fmt.Errorf(utilities.MinValueMust, strings.ToLower(label), min)
		receiver.Messages = append(receiver.Messages, map[string]interface{}{
			strings.ToLower(label): err.Error(),
		})
	}
}

// IsFloatMax :
func (receiver *Validation) IsFloatMax(value float64, max float64, label string) {
	if value > max {
		err := fmt.Errorf(utilities.MaxValueMust, strings.ToLower(label), max)
		receiver.Messages = append(receiver.Messages, map[string]interface{}{
			strings.ToLower(label): err.Error(),
		})
	}
}

// IsFloatMin :
func (receiver *Validation) IsFloatMin(value float64, min float64, label string) {
	if value < min {
		err := fmt.Errorf(utilities.MinValueMust, strings.ToLower(label), min)
		receiver.Messages = append(receiver.Messages, map[string]interface{}{
			strings.ToLower(label): err.Error(),
		})
	}
}
