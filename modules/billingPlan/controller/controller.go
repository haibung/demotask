package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/demotask/backend/models"
	"github.com/demotask/backend/modules/billingPlan"

	billingPlanRepo "github.com/demotask/backend/modules/billingPlan/repository"

	productRepo "github.com/demotask/backend/modules/product/repository"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/paypal"
	"github.com/demotask/backend/utilities"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type (
	IBillingPlanController interface {
		CreateBillingPlan(ctx context.Context, req *billingPlan.CreateBillingPlanRequest, tx *gorm.DB) (*billingPlan.CreateBillingPlanResponse, error)
		validateBillingCycles(cycles []billingPlan.CreateBillingCycleRequest) error
	}

	BillingPlanController struct {
		fx.In
		Logger                *logger.Logger
		BillingPlanRepository billingPlanRepo.IBillingPlanRepository
		ProductRepository     productRepo.IProductRepository
		PayPalClient          paypal.IPayPalClient
	}
)

func NewController(controller BillingPlanController) IBillingPlanController {
	return &controller
}

func (c *BillingPlanController) validateBillingCycles(cycles []billingPlan.CreateBillingCycleRequest) error {

	if len(cycles) == 0 {
		return errors.New("billing_cycles required")
	}

	var regularCount int
	var trialCount int

	currency := cycles[0].Currency

	for i, c := range cycles {

		if c.IntervalCount <= 0 {
			return errors.New("interval_count must be > 0")
		}

		if c.IntervalUnit != "DAY" &&
			c.IntervalUnit != "WEEK" &&
			c.IntervalUnit != "MONTH" &&
			c.IntervalUnit != "YEAR" {
			return errors.New("invalid interval_unit")
		}

		if c.Currency != currency {
			return errors.New("all billing cycles must use same currency")
		}

		if c.TenureType == "REGULAR" {
			regularCount++
		}

		if c.TenureType == "TRIAL" {
			trialCount++
			if c.TotalCycles == 0 {
				return errors.New("trial cannot be unlimited")
			}
		}

		if c.Sequence != i+1 {
			return errors.New("sequence must start from 1 and increment by 1")
		}
	}

	if regularCount != 1 {
		return errors.New("exactly one REGULAR billing cycle required")
	}

	if trialCount > 2 {
		return errors.New("maximum two TRIAL cycles allowed")
	}

	return nil
}

func (c *BillingPlanController) CreateBillingPlan(ctx context.Context, req *billingPlan.CreateBillingPlanRequest, tx *gorm.DB) (*billingPlan.CreateBillingPlanResponse, error) {

	messages, err := req.Validate()
	if err != nil {
		return nil, utilities.ErrorRequest(err, http.StatusBadRequest, messages)
	}

	if req.ProductID == nil {
		return nil, utilities.ErrorRequest(
			errors.New("product_id required"),
			http.StatusBadRequest,
		)
	}

	if err := c.validateBillingCycles(req.BillingCycles); err != nil {
		return nil, utilities.ErrorRequest(err, http.StatusBadRequest)
	}

	prod, err := c.ProductRepository.Find(ctx, &models.Product{ID: *req.ProductID})
	if err != nil {
		return nil, utilities.ErrorRequest(err, http.StatusBadRequest, "Product not found")
	}

	if prod.PaypalProductID == "" {
		return nil, utilities.ErrorRequest(
			errors.New("product not registered in Paypal"),
			http.StatusBadRequest,
		)
	}

	exists, err := c.BillingPlanRepository.ExistsActivePlan(ctx, prod.ID, req.Name)
	if err != nil {
		return nil, utilities.ErrorRequest(err, http.StatusInternalServerError)
	}

	if exists {
		return nil, utilities.ErrorRequest(
			errors.New("active plan with same name already exists"),
			http.StatusConflict,
		)
	}

	ppCycles := make([]paypal.BillingCycle, 0)

	for _, bc := range req.BillingCycles {

		ppCycles = append(ppCycles, paypal.BillingCycle{
			Frequency: paypal.Frequency{
				IntervalUnit:  bc.IntervalUnit,
				IntervalCount: bc.IntervalCount,
			},
			TenureType:  bc.TenureType,
			Sequence:    bc.Sequence,
			TotalCycles: bc.TotalCycles,
			PricingScheme: paypal.PricingScheme{
				FixedPrice: paypal.Money{
					Value:        fmt.Sprintf("%.2f", bc.Price),
					CurrencyCode: bc.Currency,
				},
			},
		})
	}

	ppReq := &paypal.CreatePlanRequest{
		ProductID:          prod.PaypalProductID,
		Name:               req.Name,
		Description:        req.Description,
		Status:             string(req.Status),
		BillingCycles:      ppCycles,
		PaymentPreferences: req.PaymentPreferences,
	}

	if req.Taxes != nil {
		ppReq.Taxes = &paypal.Taxes{
			Percentage: fmt.Sprintf("%.2f", req.Taxes.Percentage),
			Inclusive:  req.Taxes.Inclusive,
		}
	}

	ppResp, err := c.PayPalClient.CreatePlan(ppReq, utilities.GeneratePrefixedUUID("PLAN"))
	if err != nil {
		return nil, utilities.ErrorRequest(err, http.StatusBadGateway)
	}

	plan := &models.BillingPlan{
		ProductID:    prod.ID,
		Name:         req.Name,
		Description:  &req.Description,
		Status:       req.Status,
		PaypalPlanID: &ppResp.ID,
	}

	if req.PaymentPreferences != nil {
		plan.AutoBillOutstanding = req.PaymentPreferences.AutoBillOutstanding
		plan.PaymentFailureThreshold = req.PaymentPreferences.PaymentFailureThreshold
	}

	if req.Taxes != nil {
		plan.TaxPercentage = &req.Taxes.Percentage
		plan.TaxInclusive = &req.Taxes.Inclusive
	}

	if _, err := c.BillingPlanRepository.Create(ctx, plan, tx); err != nil {

		_ = c.PayPalClient.DeactivatePlan(ppResp.ID)

		return nil, utilities.ErrorRequest(err, http.StatusInternalServerError)
	}

	for _, bc := range req.BillingCycles {

		cycle := &models.BillingCycle{
			BillingPlanID: plan.ID,
			IntervalUnit:  bc.IntervalUnit,
			IntervalCount: bc.IntervalCount,
			TenureType:    bc.TenureType,
			Sequence:      bc.Sequence,
			TotalCycles:   bc.TotalCycles,
			PriceValue:    bc.Price,
			Currency:      bc.Currency,
		}

		if err := c.BillingPlanRepository.CreateBillingCycle(ctx, cycle, tx); err != nil {

			_ = c.PayPalClient.DeactivatePlan(ppResp.ID)

			return nil, utilities.ErrorRequest(err, http.StatusInternalServerError)
		}
	}

	return &billingPlan.CreateBillingPlanResponse{
		ID:           plan.ID,
		ProductID:    plan.ProductID,
		Name:         plan.Name,
		Description:  *plan.Description,
		Status:       string(plan.Status),
		PaypalPlanID: ppResp.ID,
	}, nil
}

//
// func (c *BillingPlanController) CreateBillingPlan(ctx context.Context, req *billingPlan.CreateBillingPlanRequest, tx *gorm.DB) (*billingPlan.CreateBillingPlanResponse, error) {
// 	// 1. Validate
// 	messages, err := req.Validate()
// 	if err != nil {
// 		return nil, utilities.ErrorRequest(err, http.StatusBadRequest, messages)
// 	}

// 	prod, err := c.ProductRepository.Find(ctx, &models.Product{ID: *req.ProductID})
// 	if err != nil {
// 		return nil, utilities.ErrorRequest(err, http.StatusBadRequest, "Product not found")
// 	}

// 	if prod.PaypalProductID == "" {
// 		err := errors.New("product not registered in Paypal")
// 		return nil, utilities.ErrorRequest(err, http.StatusBadRequest)
// 	}

// 	requestID := fmt.Sprintf("PLAN-%s", uuid.New().String())

// 	ppCyles := make([]paypal.BillingCycle, 0)
// 	for _, bc := range req.BillingCycles {
// 		ppCyles = append(ppCyles, paypal.BillingCycle{
// 			Frequency: paypal.Frequency{
// 				IntervalUnit:  bc.IntervalUnit,
// 				IntervalCount: bc.IntervalCount,
// 			},
// 			TenureType:  bc.TenureType,
// 			Sequence:    bc.Sequence,
// 			TotalCycles: bc.TotalCycles,
// 			PricingScheme: paypal.PricingScheme{
// 				FixedPrice: paypal.Money{
// 					Value:        fmt.Sprintf("%.2f", bc.Price),
// 					CurrencyCode: bc.Currency,
// 				},
// 			},
// 		})
// 	}

// 	ppReq := &paypal.CreatePlanRequest{
// 		ProductID:          prod.PaypalProductID,
// 		Name:               req.Name,
// 		Description:        req.Description,
// 		Status:             string(req.Status),
// 		BillingCycles:      ppCyles,
// 		PaymentPreferences: req.PaymentPreferences,
// 	}

// 	ppResp, err := c.PayPalClient.CreatePlan(ppReq, requestID)
// 	if err != nil {
// 		return nil, utilities.ErrorRequest(err, http.StatusBadGateway)
// 	}

// 	plan := &models.BillingPlan{
// 		ProductID:    prod.ID,
// 		Name:         req.Name,
// 		Description:  &req.Description,
// 		Status:       req.Status,
// 		PaypalPlanID: &ppResp.ID,
// 	}

// 	_, err = c.BillingPlanRepository.Create(ctx, plan, tx)
// 	if err != nil {
// 		return nil, utilities.ErrorRequest(err, http.StatusInternalServerError)
// 	}

// 	for _, bc := range req.BillingCycles {
// 		cycle := &models.BillingCycle{
// 			BillingPlanID: plan.ID,
// 			IntervalUnit:  bc.IntervalUnit,
// 			IntervalCount: bc.IntervalCount,
// 			TenureType:    bc.TenureType,
// 			Sequence:      bc.Sequence,
// 			TotalCycles:   bc.TotalCycles,
// 			PriceValue:    bc.Price,
// 			Currency:      bc.Currency,
// 		}
// 		if err := c.BillingPlanRepository.CreateBillingCycle(ctx, cycle, tx); err != nil {
// 			return nil, utilities.ErrorRequest(err, http.StatusInternalServerError)
// 		}
// 	}
// 	return &billingPlan.CreateBillingPlanResponse{
// 		ID:           plan.ID,
// 		ProductID:    plan.ProductID,
// 		Name:         plan.Name,
// 		Description:  *plan.Description,
// 		Status:       string(plan.Status),
// 		PaypalPlanID: ppResp.ID,
// 	}, nil
// }
