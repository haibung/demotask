package enum

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/demotask/backend/utilities"
)

type OrderStatus int

const (
	OrderStatusCreated  OrderStatus = 0
	OrderStatusApproved OrderStatus = 1
	OrderStatusCaptured OrderStatus = 2
	OrderStatusFailed   OrderStatus = 3
)

func (t OrderStatus) String() string {
	switch t {
	case OrderStatusCreated:
		return "CREATED"
	case OrderStatusApproved:
		return "APPROVED"
	case OrderStatusCaptured:
		return "CAPTURED"
	case OrderStatusFailed:
		return "FAILED"
	default:
		return "Unknown"
	}
}

func (t OrderStatus) IsValid() error {
	switch t {
	case
		OrderStatusCreated,
		OrderStatusApproved,
		OrderStatusCaptured,
		OrderStatusFailed:
		return nil
	}
	return fmt.Errorf(utilities.DataNotFound, "Status Order")
}

func OrderStatusToArray(value string) ([]int, error) {
	var result []int
	statusOrder := strings.Split(value, ",")
	for i := 0; i < len(statusOrder); i++ {
		statusOrderID, err := strconv.Atoi(statusOrder[i])
		if err != nil {
			return nil, errors.New(utilities.NumberNotValid)
		}
		if err := OrderStatus(statusOrderID).IsValid(); err != nil {
			return nil, err
		}
		result = append(result, statusOrderID)
	}
	return result, nil
}

// const (
// 	OrderStatusPending   OrderStatus = 0
// 	OrderStatusCompleted OrderStatus = 1
// 	OrderStatusFailed    OrderStatus = 2
// )

// func (t OrderStatus) String() string {
// 	switch t {
// 	case OrderStatusPending:
// 		return "PENDING"
// 	case OrderStatusCompleted:
// 		return "COMPLETED"
// 	case OrderStatusFailed:
// 		return "FAILED"
// 	default:
// 		return "Unknown"
// 	}
// }

// func (t OrderStatus) IsValid() error {
// 	switch t {
// 	case
// 		OrderStatusPending,
// 		OrderStatusCompleted,
// 		OrderStatusFailed:
// 		return nil
// 	}
// 	return fmt.Errorf(utilities.DataNotFound, "Status Order")
// }

// func OrderStatusToArray(value string) ([]int, error) {
// 	var result []int
// 	statusOrder := strings.Split(value, ",")
// 	for i := 0; i < len(statusOrder); i++ {
// 		statusOrderID, err := strconv.Atoi(statusOrder[i])
// 		if err != nil {
// 			return nil, errors.New(utilities.NumberNotValid)
// 		}
// 		if err := OrderStatus(statusOrderID).IsValid(); err != nil {
// 			return nil, err
// 		}
// 		result = append(result, statusOrderID)
// 	}
// 	return result, nil
// }
