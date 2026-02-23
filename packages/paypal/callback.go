package paypal

import (
	"encoding/json"
	"sync"
	"time"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"` // Seconds until expiry
	AppID        string `json:"app_id"`
	Nonce        string `json:"nonce"`
	Scope        string `json:"scope"`
}

// UserInfo represents PayPal user profile information
type UserInfo struct {
	UserID        string  `json:"user_id"`
	Email         string  `json:"email"`
	EmailVerified bool    `json:"verified_account"`
	Name          string  `json:"name"`
	GivenName     string  `json:"given_name"`
	FamilyName    string  `json:"family_name"`
	PayerID       string  `json:"payer_id"`
	Address       Address `json:"address"`
}

// Address represents user's address from PayPal
type Address struct {
	StreetAddress string `json:"street_address"`
	Locality      string `json:"locality"`
	Region        string `json:"region"`
	PostalCode    string `json:"postal_code"`
	Country       string `json:"country"`
}

// ErrorResponse represents PayPal API error
type ErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	Message          string `json:"message"`
	Name             string `json:"name"`
	DebugID          string `json:"debug_id"`
}

// WebhookEvent represents a PayPal webhook event
type WebhookEvent struct {
	ID           string                 `json:"id"`
	EventType    string                 `json:"event_type"`
	EventVersion string                 `json:"event_version"`
	Summary      string                 `json:"summary"`
	ResourceType string                 `json:"resource_type"`
	Resource     map[string]interface{} `json:"resource"`
	CreateTime   time.Time              `json:"create_time"`
	Links        []Link                 `json:"links"`
}

// WebhookVerificationRequest for verifying webhook signatures
type WebhookVerificationRequest struct {
	AuthAlgo         string          `json:"auth_algo"`
	CertURL          string          `json:"cert_url"`
	TransmissionID   string          `json:"transmission_id"`
	TransmissionSig  string          `json:"transmission_sig"`
	TransmissionTime string          `json:"transmission_time"`
	WebhookID        string          `json:"webhook_id"`
	WebhookEvent     json.RawMessage `json:"webhook_event"`
}

// WebhookVerificationResponse from PayPal
type WebhookVerificationResponse struct {
	VerificationStatus string `json:"verification_status"` // SUCCESS or FAILURE
}

type SubscriptionPaymentSource struct {
	PayPal *SubscriptionPaypalSource `json:"paypal"`
}

type SubscriptionPaypalSource struct {
	VaultID    string                         `json:"vault_id,omitempty"`
	Attributes *SubscriptionPaymentAttributes `json:"attributes,omitempty"`
}

type SubscriptionPaymentAttributes struct {
	Vault *SubscriptionVault `json:"vault,omitempty"`
}

type SubscriptionVault struct {
	StoreInVault string `json:"store_in_vault,omitempty"`
	UsageType    string `json:"usage_type,omitempty"`
}
type CreateSubscriptionRequest struct {
	PlanID             string                     `json:"plan_id"`
	Quantity           int                        `json:"quantity"`
	StartTime          *string                    `json:"start_time"`
	Subscriber         *SubscriptionSubscriber    `json:"subscriber"`
	ApplicationContext *ApplicationContext        `json:"application_context"`
	PaymentSource      *SubscriptionPaymentSource `json:"payment_source"`
}

type SubscriptionSubscriber struct {
	Name         *SubscriptionName `json:"name"`
	EmailAddress string            `json:"email_address"`
}

type SubscriptionName struct {
	GivenName string `json:"given_name"`
	Surname   string `json:"surname"`
}

type CreateSubscriptionResponse struct {
	ID               string `json:"id"`
	Status           string `json:"status"`
	StatusUpdateTime string `json:"status_update_time,omitempty"`
	PlanID           string `json:"plan_id"`
	StartTime        string `json:"start_time,omitempty"`
	Quantity         string `json:"quantity,omitempty"`
	CreateTime       string `json:"create_time"`
	UpdateTime       string `json:"update_time,omitempty"`
	Links            []Link `json:"links"`
}

type VaultSetupTokenRequest struct {
	PaymentSource VaultSetupPaymentSource `json:"payment_source"`
}

type VaultSetupPaymentSource struct {
	PayPal *VaultPayPalSource `json:"paypal,omitempty"`
	Card   *VaultCardSource   `json:"card,omitempty"`
}

type VaultPayPalSource struct {
	ExperienceContext *VaultExperienceContext `json:"experience_context,omitempty"`
}

type VaultExperienceContext struct {
	ReturnURL string `json:"return_url,omitempty"`
	CancelURL string `json:"cancel_url,omitempty"`
}

type VaultSetupTokenResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Links  []Link `json:"links"`
}

type CreatePaymentTokenRequest struct {
	PaymentSource VaultPaymentTokenSource `json:"payment_source"`
}

type VaultPaymentTokenSource struct {
	Token VaultToken `json:"token"`
}

type VaultToken struct {
	ID   string `json:"id"`
	Type string `json:"type"` // SETUP_TOKEN
}
type CreatePaymentTokenResponse struct {
	ID            string                `json:"id"`
	Customer      *VaultCustomer        `json:"customer,omitempty"`
	PaymentSource *VaultedPaymentSource `json:"payment_source,omitempty"`
	Links         []Link                `json:"links"`
}

type VaultCardSource struct {
	Number         string          `json:"number,omitempty"`
	Expiry         string          `json:"expiry,omitempty"` // YYYY-MM
	SecurityCode   string          `json:"security_code,omitempty"`
	Name           string          `json:"name,omitempty"`
	BillingAddress *BillingAddress `json:"billing_address,omitempty"`
}

type BillingAddress struct {
	AddressLine1 string `json:"address_line_1,omitempty"`
	AddressLine2 string `json:"address_line_2,omitempty"`
	AdminArea1   string `json:"admin_area_1,omitempty"`
	AdminArea2   string `json:"admin_area_2,omitempty"`
	PostalCode   string `json:"postal_code,omitempty"`
	CountryCode  string `json:"country_code,omitempty"`
}

type VaultCustomer struct {
	ID string `json:"id"`
}

type VaultedPaymentSource struct {
	PayPal *VaultedPayPal `json:"paypal,omitempty"`
	Card   *VaultedCard   `json:"card,omitempty"`
}

type VaultedPayPal struct {
	EmailAddress string `json:"email_address,omitempty"`
	AccountID    string `json:"account_id,omitempty"`
}

type VaultedCard struct {
	Brand          string `json:"brand,omitempty"`
	LastDigits     string `json:"last_digits,omitempty"`
	ExpirationDate string `json:"expiration_date,omitempty"` // YYYY-MM
}

type CreateProductRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type"` // SERVICE | DIGITAL | PHYSICAL
	Category    string `json:"category,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
	HomeURL     string `json:"home_url,omitempty"`
}

type CreateProductResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`
	Category    string `json:"category,omitempty"`
	CreateTime  string `json:"create_time,omitempty"`
	Links       []Link `json:"links"`
}

type CreatePlanRequest struct {
	ProductID          string              `json:"product_id"`
	Name               string              `json:"name"`
	Description        string              `json:"description,omitempty"`
	Status             string              `json:"status"` // ACTIVE | CREATED
	BillingCycles      []BillingCycle      `json:"billing_cycles"`
	PaymentPreferences *PaymentPreferences `json:"payment_preferences,omitempty"`
	Taxes              *Taxes              `json:"taxes,omitempty"`
}

type BillingCycle struct {
	Frequency     Frequency     `json:"frequency"`
	TenureType    string        `json:"tenure_type"` // REGULAR | TRIAL
	Sequence      int           `json:"sequence"`
	TotalCycles   int           `json:"total_cycles"`
	PricingScheme PricingScheme `json:"pricing_scheme"`
}

type Frequency struct {
	IntervalUnit  string `json:"interval_unit"` // DAY | WEEK | MONTH | YEAR
	IntervalCount int    `json:"interval_count"`
}

type PricingScheme struct {
	FixedPrice Money `json:"fixed_price"`
}

type PaymentPreferences struct {
	AutoBillOutstanding     bool `json:"auto_bill_outstanding"`
	PaymentFailureThreshold int  `json:"payment_failure_threshold,omitempty"`
}

type Taxes struct {
	Percentage string `json:"percentage"`
	Inclusive  bool   `json:"inclusive"`
}

type CreatePlanResponse struct {
	ID            string         `json:"id"`
	ProductID     string         `json:"product_id"`
	Name          string         `json:"name"`
	Status        string         `json:"status"`
	CreateTime    string         `json:"create_time"`
	UpdateTime    string         `json:"update_time"`
	Links         []Link         `json:"links"`
	BillingCycles []BillingCycle `json:"billing_cycles,omitempty"`
}

type CreateOrderResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type VaultTokenDetailsResponse struct {
	ID            string                   `json:"id"`
	Status        string                   `json:"status"`
	Customer      *VaultCustomer           `json:"customer,omitempty"`
	PaymentSource *VaultTokenPaymentSource `json:"payment_source,omitempty"`
	CreateTime    string                   `json:"create_time,omitempty"`
}

type VaultTokenPaymentSource struct {
	PayPal *VaultTokenPayPal `json:"paypal,omitempty"`
}

type VaultTokenPayPal struct {
	EmailAddress string `json:"email_address,omitempty"`
	PayerID      string `json:"payer_id,omitempty"`
	UsageType    string `json:"usage_type,omitempty"`
	CustomerType string `json:"customer_type,omitempty"`
}
type PaypalClient struct {
	AccessToken string
	Expiry      time.Time
	Mutex       sync.Mutex
}

// type VaultOrderRequest struct {
// 	Intent        string                  `json:"intent"`
// 	PurchaseUnits []VaultPurchaseUnit     `json:"purchase_units"`
// 	PaymentSource VaultOrderPaymentSource `json:"payment_source"`
// }

type VaultOrderPaymentSource struct {
	PayPal VaultOrderPayPal `json:"paypal"`
}

type VaultOrderPayPal struct {
	VaultID string `json:"vault_id"`
}

type CheckoutOrderResource struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Intent string `json:"intent"`
}

type Money struct {
	CurrencyCode string `json:"currency_code"`
	Value        string `json:"value"`
}

type Link struct {
	Href   string `json:"href"`
	Rel    string `json:"rel"`
	Method string `json:"method"`
}

type OrderRequest struct {
	Intent             string              `json:"intent"`
	PurchaseUnits      []PurchaseUnit      `json:"purchase_units"`
	PaymentSource      *PaymentSource      `json:"payment_source,omitempty"`
	ApplicationContext *ApplicationContext `json:"application_context,omitempty"`
}

type PurchaseUnit struct {
	ReferenceID string        `json:"reference_id,omitempty"`
	Amount      Amount        `json:"amount"`
	Description string        `json:"description,omitempty"`
	Items       []Item        `json:"items,omitempty"`
	Payments    *UnitPayments `json:"payments,omitempty"` // ONLY populated in capture response
}

type Amount struct {
	CurrencyCode string           `json:"currency_code"`
	Value        string           `json:"value"`
	Breakdown    *AmountBreakdown `json:"breakdown,omitempty"`
}

type AmountBreakdown struct {
	ItemTotal *Money `json:"item_total,omitempty"`
}

type Item struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Quantity    string `json:"quantity"`
	UnitAmount  Money  `json:"unit_amount"`
}

type OrderResponse struct {
	ID            string         `json:"id"`
	Status        string         `json:"status"`
	Links         []Link         `json:"links"`
	PurchaseUnits []PurchaseUnit `json:"purchase_units"`
}

type CaptureResponse struct {
	ID            string         `json:"id"`
	Status        string         `json:"status"`
	PurchaseUnits []PurchaseUnit `json:"purchase_units"`
	Payer         *Payer         `json:"payer,omitempty"`
	PaymentSource *PaymentSource `json:"payment_source,omitempty"`
}

type UnitPayments struct {
	Captures []Capture `json:"captures"`
}

type Capture struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	Amount       Money  `json:"amount"`
	FinalCapture bool   `json:"final_capture"`
	CreateTime   string `json:"create_time"`
	UpdateTime   string `json:"update_time"`
}

type Payer struct {
	EmailAddress string `json:"email_address"`
	PayerID      string `json:"payer_id"`
}

type PaymentSource struct {
	Token  *PaymentToken        `json:"token,omitempty"`
	PayPal *PayPalAccountSource `json:"paypal,omitempty"`
}

type PayPalAccountSource struct {
	EmailAddress  string                   `json:"email_address,omitempty"`
	AccountID     string                   `json:"account_id,omitempty"`
	AccountStatus string                   `json:"account_status,omitempty"`
	Attributes    *PaymentSourceAttributes `json:"attributes,omitempty"`
	Vault         *PayPalVaultResponse     `json:"vault,omitempty"`
}

type PaymentToken struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type ApplicationContext struct {
	BrandName          string `json:"brand_name,omitempty"`
	Locale             string `json:"locale,omitempty"`
	ShippingPreference string `json:"shipping_preference,omitempty"`
	UserAction         string `json:"user_action,omitempty"`
	ReturnURL          string `json:"return_url,omitempty"`
	CancelURL          string `json:"cancel_url,omitempty"`
}

type PaymentSourceAttributes struct {
	Vault *PaymentSourceVault `json:"vault,omitempty"`
}

type PaymentSourceVault struct {
	StoreInVault string `json:"store_in_vault,omitempty"`
	UsageType    string `json:"usage_type,omitempty"`
}

type PayPalVaultResponse struct {
	ID     string `json:"id,omitempty"`
	Status string `json:"status,omitempty"`
}

///////////

type VaultOrderRequest struct {
	Intent        string              `json:"intent"`
	PurchaseUnits []VaultPurchaseUnit `json:"purchase_units"`
	PaymentSource VaultPaymentSource  `json:"payment_source"`
}

type VaultPurchaseUnit struct {
	Amount VaultAmount `json:"amount"`
}

type VaultAmount struct {
	CurrencyCode string `json:"currency_code"`
	Value        string `json:"value"`
}

type VaultPaymentSource struct {
	PayPal PaypalSource `json:"paypal"`
}

type PaypalSource struct {
	VaultID string `json:"vault_id"`
}

// type SubscriptionDetailResponse struct {
// 	ID string `json:"id"`

// 	PaymentSource *struct {
// 		PayPal *struct {
// 			VaultID string `json:"vault_id"`
// 			Email   string `json:"email_address"`
// 		} `json:"paypal"`
// 	} `json:"payment_source"`
// }

type SubscriptionDetailResponse struct {
	ID            string       `json:"id"`
	Status        string       `json:"status"`
	StartTime     string       `json:"start_time"`
	Quantity      string       `json:"quantity"`
	BillingInfo   *BillingInfo `json:"billing_info"`
	PaymentSource *struct {
		PayPal *struct {
			VaultID string `json:"vault_id"`
			Email   string `json:"email_address"`
		} `json:"paypal"`
	} `json:"payment_source"`
}

type BillingInfo struct {
	NextBillingTime     string       `json:"next_billing_time"`
	FailedPaymentsCount int          `json:"failed_payments_count"`
	LastPayment         *LastPayment `json:"last_payment"`
}

type LastPayment struct {
	Time string `json:"time"`
}
