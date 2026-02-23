package enum

type OrderStatusInternal string

const (
	OrderStatusInternalPending        OrderStatusInternal = "PENDING"
	OrderStatusInternalWaitingPayment OrderStatusInternal = "WAITING_PAYMENT"
	OrderStatusInternalPaid           OrderStatusInternal = "PAID"
	OrderStatusInternalCancelled      OrderStatusInternal = "CANCELLED"
	OrderStatusInternalFailed         OrderStatusInternal = "FAILED"
)

func (s OrderStatusInternal) String() string {
	return string(s)
}

func (s OrderStatusInternal) IsValid() bool {
	switch s {
	case
		OrderStatusInternalPending,
		OrderStatusInternalWaitingPayment,
		OrderStatusInternalPaid,
		OrderStatusInternalCancelled,
		OrderStatusInternalFailed:
		return true
	default:
		return false
	}
}
