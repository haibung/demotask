package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/demotask/backend/enum"
	"github.com/demotask/backend/models"
	"github.com/demotask/backend/modules/product"
	"github.com/demotask/backend/modules/product/repository"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/paypal"
	"github.com/demotask/backend/utilities"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type (
	IProductController interface {
		CreateProduct(ctx context.Context, req *product.CreateProductRequest, tx *gorm.DB) (*product.ProductResponse, error)
		CreateOneTimePrice(ctx context.Context, req *product.CreateOneTimePriceRequest, tx *gorm.DB) (*product.CreateOneTimePriceResponse, error)
		// CreatePlan(ctx context.Context, req *product.CreatePlanRequest, tx *gorm.DB) (*product.CreatePlanResponse, error)
		GetProducts(ctx context.Context, req *product.GetProductRequest) ([]product.ProductPricingResponse, error)
	}

	ProductController struct {
		fx.In
		Logger     *logger.Logger
		Repository repository.IProductRepository
		// BillingPlanRepo repository.IBillingPlanRepository
		PayPalClient paypal.IPayPalClient
	}
)

func NewController(controller ProductController) IProductController {
	return &controller
}

// TODO: Product must to be can identify, where product is one time or subscription
func (c *ProductController) CreateProduct(ctx context.Context, req *product.CreateProductRequest, tx *gorm.DB) (*product.ProductResponse, error) {
	// Validate request
	messages, err := req.Validate()
	if err != nil {
		c.Logger.Error(err)
		return nil, utilities.ErrorRequest(err, http.StatusBadRequest, messages)
	}

	// Create product to PayPal
	paypalProductReq := &paypal.CreateProductRequest{
		Name:        req.Name,
		Description: req.Description,
		Type:        string(req.BusinessType),
		Category:    req.Category,
	}

	if req.ImageURL != "" {
		paypalProductReq.ImageURL = req.ImageURL
	}
	if req.HomeURL != "" {
		paypalProductReq.HomeURL = req.HomeURL
	}

	paypalResp, err := c.PayPalClient.CreateProduct(paypalProductReq)
	if err != nil {
		c.Logger.Error(err)
		return nil, utilities.ErrorRequest(err, http.StatusInternalServerError)
	}

	// Create Product Model with PayPal product ID
	newProduct := &models.Product{
		Name:            req.Name,
		Description:     req.Description,
		BusinessType:    req.BusinessType,
		PaypalProductID: paypalResp.ID,
		Category:        req.Category,
		ImageURL:        req.ImageURL,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	id, err := c.Repository.Create(ctx, newProduct, tx)
	if err != nil {
		c.Logger.Error(err)
		return nil, utilities.ErrorRequest(err, http.StatusInternalServerError)
	}

	return &product.ProductResponse{
		ID:              int(*id),
		PaypalProductID: &newProduct.PaypalProductID,
		Name:            newProduct.Name,
		Description:     newProduct.Description,
		BusinessType:    newProduct.BusinessType,
		Category:        newProduct.Category,
		ImageURL:        newProduct.ImageURL,
	}, nil
}

func (c *ProductController) CreateOneTimePrice(ctx context.Context, req *product.CreateOneTimePriceRequest, tx *gorm.DB) (*product.CreateOneTimePriceResponse, error) {

	// 1. Validate
	messages, err := req.Validate()
	if err != nil {
		return nil, utilities.ErrorRequest(err, http.StatusBadRequest, messages)
	}

	productModel, err := c.Repository.Find(ctx, &models.Product{
		ID: int(req.ProductID),
	})
	if err != nil {
		c.Logger.Error(err)
		return nil, utilities.ErrorRequest(err, http.StatusNotFound)
	}

	price := &models.ProductOneTimePrice{
		ProductID: productModel.ID,
		Price:     req.Price,
		Currency:  req.Currency,
		IsActive:  true,
	}

	_, err = c.Repository.CreateOneTimePrice(ctx, price, tx)
	if err != nil {
		return nil, utilities.ErrorRequest(err, http.StatusInternalServerError)
	}

	return &product.CreateOneTimePriceResponse{
		ID:        price.ID,
		ProductID: price.ProductID,
		Price:     price.Price,
		Currency:  price.Currency,
		IsActive:  price.IsActive,
		CreatedAt: price.CreatedAt,
	}, nil
}

func (c *ProductController) GetProducts(ctx context.Context, req *product.GetProductRequest) ([]product.ProductPricingResponse, error) {

	products, err := c.Repository.FindAll(ctx, nil)
	if err != nil {
		return nil, utilities.ErrorRequest(err, http.StatusInternalServerError)
	}

	var responses []product.ProductPricingResponse

	for _, p := range products {

		resp := product.ProductPricingResponse{
			Product: product.ProductResponse{
				ID:           int(p.ID),
				Name:         p.Name,
				Description:  p.Description,
				BusinessType: p.BusinessType,
				Category:     p.Category,
				ImageURL:     p.ImageURL,
				HomeURL:      p.HomeURL,
				CreatedAt:    p.CreatedAt,
				UpdatedAt:    p.UpdatedAt,
			},
		}

		var pricingTypes []string

		if len(p.OneTimePrices) > 0 {

			pricingTypes = append(pricingTypes, "ONE_TIME")

			activePrice := p.OneTimePrices[0]

			resp.OneTimePrice = &product.OneTimePriceResponse{
				Price:    activePrice.Price,
				Currency: activePrice.Currency,
			}
		}

		if len(p.BillingPlans) > 0 {

			var subOptions []product.SubscriptionOptionResponse

			for _, plan := range p.BillingPlans {

				if plan.Status != enum.PlanStatusActive {
					continue
				}

				var regularCycle *models.BillingCycle
				var trialCycle *models.BillingCycle

				for i := range plan.BillingCycles {
					cycle := plan.BillingCycles[i]
					switch cycle.TenureType {
					case "REGULAR":
						regularCycle = &cycle
					case "TRIAL":
						trialCycle = &cycle
					}
				}

				if regularCycle == nil {
					continue
				}

				basePrice := regularCycle.PriceValue
				finalPrice := basePrice

				if plan.TaxPercentage != nil && plan.TaxInclusive != nil {
					taxPercent := *plan.TaxPercentage
					if *plan.TaxInclusive {
						finalPrice = basePrice
					} else {
						finalPrice = basePrice + (basePrice * taxPercent / 100)
					}
				}

				option := product.SubscriptionOptionResponse{
					BillingPlanID: int(plan.ID),
					IntervalUnit:  regularCycle.IntervalUnit,
					IntervalCount: regularCycle.IntervalCount,
					BasePrice:     basePrice,
					FinalPrice:    finalPrice,
					Currency:      regularCycle.Currency,
					TaxPercentage: plan.TaxPercentage,
					TaxInclusive:  plan.TaxInclusive,
				}

				if trialCycle != nil {
					option.HasTrial = true
					option.TrialIntervalUnit = trialCycle.IntervalUnit
					option.TrialIntervalCount = trialCycle.IntervalCount
					option.TrialCycles = trialCycle.TotalCycles
				}

				subOptions = append(subOptions, option)
			}

			// for _, cycle := range plan.BillingCycles {

			// 	if cycle.TenureType != "REGULAR" {
			// 		continue
			// 	}

			// subOptions = append(subOptions,
			// 	product.SubscriptionOptionResponse{
			// 		BillingPlanID: int(plan.ID),
			// 		IntervalUnit:  cycle.IntervalUnit,
			// 		IntervalCount: cycle.IntervalCount,
			// 		Price:         cycle.PriceValue,
			// 		Currency:      cycle.Currency,
			// 	},
			// )
			// }
			// }

			if len(subOptions) > 0 {
				pricingTypes = append(pricingTypes, "SUBSCRIPTION")
				resp.SubscriptionOptions = subOptions
			}
		}

		if len(pricingTypes) > 0 {
			resp.PricingTypes = pricingTypes
		}

		responses = append(responses, resp)
	}

	return responses, nil
}

// func (c *ProductController) GetProducts(ctx context.Context, req *product.GetProductRequest) ([]product.ProductPricingResponse, error) {

// 	products, err := c.Repository.FindAll(ctx, nil)
// 	if err != nil {
// 		return nil, utilities.ErrorRequest(err, http.StatusInternalServerError)
// 	}

// 	var responses []product.ProductPricingResponse

// 	for _, p := range products {

// 		resp := product.ProductPricingResponse{
// 			Product: product.ProductResponse{
// 				ID:           int64(p.ID),
// 				Name:         p.Name,
// 				Description:  p.Description,
// 				BusinessType: p.BusinessType,
// 				Category:     p.Category,
// 				ImageURL:     p.ImageURL,
// 				HomeURL:      p.HomeURL,
// 				CreatedAt:    p.CreatedAt,
// 				UpdatedAt:    p.UpdatedAt,
// 			},
// 		}

// 		// One time pricing
// 		if len(p.OneTimePrices) > 0 {
// 			resp.OneTimePrice = &struct {
// 				Price    float64 `json:"price"`
// 				Currency string  `json:"currency"`
// 			}{
// 				Price:    p.OneTimePrices[0].Price,
// 				Currency: p.OneTimePrices[0].Currency,
// 			}
// 		}

// 		// Subscription options
// 		if len(p.BillingPlans) > 0 {
// 			for _, plan := range p.BillingPlans {

// 				for _, cycle := range plan.BillingCycles {

// 					if cycle.TenureType != "REGULAR" {
// 						continue
// 					}

// 					resp.SubscriptionOptions = append(resp.SubscriptionOptions, struct {
// 						BillingPlanID int     `json:"billing_plan_id"`
// 						IntervalUnit  string  `json:"interval_unit"`
// 						IntervalCount int     `json:"interval_count"`
// 						Price         float64 `json:"price"`
// 						Currency      string  `json:"currency"`
// 					}{
// 						BillingPlanID: int(plan.ID),
// 						IntervalUnit:  cycle.IntervalUnit,
// 						IntervalCount: cycle.IntervalCount,
// 						Price:         cycle.PriceValue,
// 						Currency:      cycle.Currency,
// 					})
// 				}
// 			}
// 		}

// 		responses = append(responses, resp)
// 	}

// 	return responses, nil
// }
