package enum

type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "PENDING"
	PaymentStatusCaptured PaymentStatus = "CAPTURED"
	PaymentStatusFailed   PaymentStatus = "FAILED"
	PaymentStatusRefunded PaymentStatus = "REFUNDED"
)

func (s PaymentStatus) String() string {
	return string(s)
}

func (s PaymentStatus) IsValid() bool {
	switch s {
	case PaymentStatusPending,
		PaymentStatusCaptured,
		PaymentStatusFailed,
		PaymentStatusRefunded:
		return true
	default:
		return false
	}
}

type PaymentMethodType string

const (
	PaymentMethodPaypal PaymentMethodType = "PAYPAL"
)
