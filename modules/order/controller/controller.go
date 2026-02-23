package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/demotask/backend/config"
	"github.com/demotask/backend/enum"
	"github.com/demotask/backend/models"
	"github.com/demotask/backend/modules/order"
	"github.com/demotask/backend/modules/order/repository"
	paymentRepository "github.com/demotask/backend/modules/payment/repository"
	productRepository "github.com/demotask/backend/modules/product/repository"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/paypal"
	"github.com/demotask/backend/utilities"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type (
	IOrderController interface {
		CreateOrder(ctx context.Context, req *order.CreateOrderRequest, tx *gorm.DB) (*order.CreateOrderResponse, error)
		PayOrder(ctx context.Context, req *order.PayOrderRequest, tx *gorm.DB) (*order.PayOrderResponse, error)
		GetOrderByID(ctx context.Context, orderID int, userID int) (*order.GetOrderResponse, error)
		GetAllOrders(ctx context.Context, userID int) ([]order.GetMyOrdersResponse, error)
	}

	OrderController struct {
		fx.In
		Config       *config.Config
		Logger       *logger.Logger
		OrderRepo    repository.IOrderRepository
		ProductRepo  productRepository.IProductRepository
		PaymentRepo  paymentRepository.IPaymentRepository
		PayPalClient paypal.IPayPalClient
	}
)

func NewController(controller OrderController) IOrderController {
	return &controller
}
func (c *OrderController) CreateOrder(ctx context.Context, req *order.CreateOrderRequest, tx *gorm.DB) (*order.CreateOrderResponse, error) {

	messages, err := req.Validate()
	if err != nil {
		return nil, utilities.ErrorRequest(err, http.StatusBadRequest, messages)
	}

	userID := *req.ContextUserID

	productIDs := make([]int, 0, len(req.Items))
	for _, item := range req.Items {
		productIDs = append(productIDs, item.ProductID)
	}

	products, err := c.ProductRepo.FindByIDs(ctx, productIDs)
	if err != nil {
		c.Logger.Error(err)
		return nil, utilities.ErrorRequest(err, http.StatusInternalServerError)
	}

	productMap := make(map[int]*models.Product)
	for i := range products {
		productMap[products[i].ID] = &products[i]
	}

	for _, item := range req.Items {
		if _, ok := productMap[item.ProductID]; !ok {
			return nil, utilities.ErrorRequest(
				fmt.Errorf("product not found: %d", item.ProductID),
				http.StatusBadRequest,
			)
		}
	}

	var (
		snapshotItems []order.OrderItemSnapshot
		totalValue    float64
		currency      = "USD"
	)

	for _, item := range req.Items {
		product := productMap[item.ProductID]

		if len(product.OneTimePrices) == 0 {
			return nil, utilities.ErrorRequest(
				fmt.Errorf("product %s has no active price", product.Name),
				http.StatusBadRequest,
			)
		}
		priceVal := product.OneTimePrices[0].Price
		prodCurrency := product.OneTimePrices[0].Currency

		if currency == "USD" && prodCurrency != "" {
			currency = prodCurrency
		}

		lineTotal := priceVal * float64(item.Quantity)

		snapshotItems = append(snapshotItems, order.OrderItemSnapshot{
			ProductID:   product.ID,
			ProductName: product.Name,
			Quantity:    item.Quantity,
			UnitPrice:   priceVal,
			TotalPrice:  lineTotal,
			Currency:    currency,
		})

		totalValue += lineTotal
	}

	orderSnapshot := order.OrderSnapshot{
		Items:      snapshotItems,
		TotalValue: totalValue,
		Currency:   currency,
	}

	snapshotJSON, err := json.Marshal(orderSnapshot)
	if err != nil {
		c.Logger.Error(err)
		return nil, utilities.ErrorRequest(err, http.StatusInternalServerError)
	}

	orderModel := &models.Order{
		UserID:    userID,
		Status:    enum.OrderStatusInternalPending,
		Snapshot:  string(snapshotJSON),
		OrderType: enum.OrderTypeOneTime,
	}

	orderID, err := c.OrderRepo.CreateOrder(ctx, orderModel, tx)
	if err != nil {
		c.Logger.Error(err)
		return nil, utilities.ErrorRequest(err, http.StatusInternalServerError)
	}

	var orderItems []models.OrderItem
	for _, snap := range snapshotItems {
		orderItems = append(orderItems, models.OrderItem{
			OrderID:     *orderID,
			ProductID:   snap.ProductID,
			ProductName: snap.ProductName,
			Quantity:    snap.Quantity,
			UnitPrice:   snap.UnitPrice,
			TotalPrice:  snap.TotalPrice,
			Currency:    snap.Currency,
		})
	}

	if err := c.OrderRepo.CreateOrderItems(ctx, orderItems, tx); err != nil {
		c.Logger.Error(err)
		return nil, utilities.ErrorRequest(err, http.StatusInternalServerError)
	}

	return &order.CreateOrderResponse{
		OrderID: fmt.Sprintf("%d", *orderID),
		Status:  enum.OrderStatusInternalPending.String(),
	}, nil
}

func (c *OrderController) GetOrderByID(ctx context.Context, orderID int, userID int) (*order.GetOrderResponse, error) {

	orderModel, paymentModel, err := c.OrderRepo.FindByIDWithPayment(ctx, orderID)
	if err != nil {
		return nil, utilities.ErrorRequest(err, http.StatusNotFound)
	}

	if orderModel.UserID != userID {
		return nil, utilities.ErrorRequest(
			errors.New("forbidden"),
			http.StatusForbidden,
		)
	}

	var total float64
	var currency string

	items := make([]order.OrderItemResponse, 0)

	for _, item := range orderModel.Items {
		total += item.TotalPrice
		currency = item.Currency

		items = append(items, order.OrderItemResponse{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			TotalPrice:  item.TotalPrice,
			Currency:    item.Currency,
		})
	}

	var paymentInfo *order.PaymentInfo

	if paymentModel != nil {
		paymentInfo = &order.PaymentInfo{
			Provider:        paymentModel.Provider,
			PaypalOrderID:   utilities.DereferenceString(paymentModel.PaypalOrderID),
			PaypalCaptureID: utilities.DereferenceString(paymentModel.PaypalCaptureID),
			Status:          paymentModel.Status.String(),
			PaidAt:          paymentModel.PaidAt,
		}
	}

	return &order.GetOrderResponse{
		OrderID:    orderModel.ID,
		Status:     orderModel.Status.String(),
		TotalValue: total,
		Currency:   currency,
		Items:      items,
		Payment:    paymentInfo,
		CreatedAt:  orderModel.CreatedAt,
	}, nil
}

func (c *OrderController) GetAllOrders(ctx context.Context, userID int) ([]order.GetMyOrdersResponse, error) {
	orders, err := c.OrderRepo.FindAllWithRelationsByUserID(ctx, userID)
	if err != nil {
		return nil, utilities.ErrorRequest(err, http.StatusInternalServerError)
	}

	var responses []order.GetMyOrdersResponse

	for _, o := range orders {
		response := order.GetMyOrdersResponse{
			OrderID:       o.ID,
			PaypalOrderID: utilities.DereferenceString(o.PaypalOrderID),
			OrderType:     o.OrderType.String(),
			Status:        o.Status.String(),
			CreatedAt:     o.CreatedAt,
		}

		// ONE TIME
		if o.OrderType == enum.OrderTypeOneTime {
			var total float64
			var currency string
			var items []order.OrderItemDetail

			for _, item := range o.Items {
				total += item.TotalPrice
				currency = item.Currency

				items = append(items, order.OrderItemDetail{
					ProductID:   item.ProductID,
					ProductName: item.ProductName,
					Quantity:    item.Quantity,
					UnitPrice:   item.UnitPrice,
					TotalPrice:  item.TotalPrice,
					Currency:    item.Currency,
				})
			}

			response.TotalValue = total
			response.Currency = currency
			response.Items = items
		}

		// SUBSCRIPTION
		if o.OrderType == enum.OrderTypeSubscription && o.Subscription != nil {

			sub := o.Subscription
			plan := sub.BillingPlan

			var intervalUnit string
			var intervalCount int
			var priceValue float64
			var currency string

			if plan != nil && len(plan.BillingCycles) > 0 {
				cycle := plan.BillingCycles[0]
				intervalUnit = cycle.IntervalUnit
				intervalCount = cycle.IntervalCount
				priceValue = cycle.PriceValue
				currency = cycle.Currency
			}

			response.Subscription = &order.SubscriptionDetail{
				SubscriptionID:       sub.ID,
				PaypalSubscriptionID: sub.PaypalSubscriptionID,
				Status:               sub.Status.String(),
				PlanName:             plan.Name,
				IntervalUnit:         intervalUnit,
				IntervalCount:        intervalCount,
				PriceValue:           priceValue,
				Currency:             currency,
				StartTime:            sub.StartTime,
				NextBillingTime:      sub.NextBillingTime,
				FailedPaymentsCount:  sub.FailedPaymentsCount,
				LastPaymentTime:      sub.LastPaymentTime,
			}

			response.TotalValue = priceValue
			response.Currency = currency
		}

		responses = append(responses, response)
	}

	return responses, nil
}

func (c *OrderController) PayOrder(ctx context.Context, req *order.PayOrderRequest, tx *gorm.DB) (*order.PayOrderResponse, error) {

	// 1. Validate request
	messages, err := req.Validate()
	if err != nil {
		return nil, utilities.ErrorRequest(err, http.StatusBadRequest, messages)
	}

	// 2. Load order
	ord, err := c.OrderRepo.FindOrderID(ctx, &models.Order{
		ID: req.OrderID,
	})
	if err != nil {
		return nil, utilities.ErrorRequest(err, http.StatusNotFound)
	}

	// 3. Ensure order belongs to user
	if ord.UserID != *req.ContextUserID {
		return nil, utilities.ErrorRequest(
			errors.New("forbidden"),
			http.StatusForbidden,
		)
	}

	// 4. Prevent duplicate PayPal order creation
	if ord.PaypalOrderID != nil {
		return nil, utilities.ErrorRequest(
			errors.New("order already sent to paypal"),
			http.StatusConflict,
		)
	}

	// 5. Deserialize snapshot
	var snap order.OrderSnapshot
	if err := json.Unmarshal([]byte(ord.Snapshot), &snap); err != nil {
		c.Logger.Error(err)
		return nil, utilities.ErrorRequest(err, http.StatusInternalServerError)
	}

	// 6. Map items to PayPal format
	var paypalItems []paypal.Item
	for _, it := range snap.Items {
		paypalItems = append(paypalItems, paypal.Item{
			Name:     it.ProductName,
			Quantity: fmt.Sprintf("%d", it.Quantity),
			UnitAmount: paypal.Money{
				CurrencyCode: it.Currency,
				Value:        fmt.Sprintf("%.2f", it.UnitPrice),
			},
		})
	}

	totalStr := fmt.Sprintf("%.2f", snap.TotalValue)

	// 7. Build PayPal order request
	ppReq := &paypal.OrderRequest{
		Intent: "CAPTURE",

		PurchaseUnits: []paypal.PurchaseUnit{
			{
				Amount: paypal.Amount{
					CurrencyCode: snap.Currency,
					Value:        totalStr,
					Breakdown: &paypal.AmountBreakdown{
						ItemTotal: &paypal.Money{
							CurrencyCode: snap.Currency,
							Value:        totalStr,
						},
					},
				},
				Items: paypalItems,
			},
		},

		ApplicationContext: &paypal.ApplicationContext{
			UserAction: "PAY_NOW",
			ReturnURL:  c.Config.PayPal.CheckoutSuccessURL,
			CancelURL:  c.Config.PayPal.CheckoutCancelURL,
		},

		PaymentSource: &paypal.PaymentSource{
			PayPal: &paypal.PayPalAccountSource{
				Attributes: &paypal.PaymentSourceAttributes{
					Vault: &paypal.PaymentSourceVault{
						StoreInVault: "ON_SUCCESS",
						UsageType:    "MERCHANT",
					},
				},
			},
		},
	}

	// 8. Call PayPal
	ppResp, err := c.PayPalClient.CreateOrder(ppReq)
	if err != nil {
		c.Logger.Error(err)
		return nil, utilities.ErrorRequest(err, http.StatusBadGateway)
	}

	// 9. Extract approval URL
	var approvalURL string
	for _, l := range ppResp.Links {
		if l.Rel == "approve" || l.Rel == "payer-action" {
			approvalURL = l.Href
			break
		}
	}

	if approvalURL == "" {
		return nil, utilities.ErrorRequest(
			errors.New("paypal approval url not found"),
			http.StatusBadGateway,
		)
	}

	// 10. Update order with PayPal ID and status
	if err := c.OrderRepo.UpdatePaypalOrderID(ctx, ord.ID, ppResp.ID, tx); err != nil {
		return nil, utilities.ErrorRequest(err, http.StatusInternalServerError)
	}

	if err := c.OrderRepo.UpdateStatus(ctx, ord.ID, enum.OrderStatusInternalWaitingPayment, tx); err != nil {
		return nil, utilities.ErrorRequest(err, http.StatusInternalServerError)
	}

	// 11. Return response
	return &order.PayOrderResponse{
		OrderID:       ord.ID,
		PaypalOrderID: ppResp.ID,
		Status:        enum.OrderStatusInternalWaitingPayment.String(),
		ApprovalURL:   approvalURL,
	}, nil
}
