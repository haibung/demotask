package route

import (
	"net/http"

	"github.com/demotask/backend/modules/product"
	"github.com/demotask/backend/modules/product/controller"
	"github.com/demotask/backend/packages/jwt"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/postgres"
	"github.com/demotask/backend/routers"
	"github.com/demotask/backend/utilities"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

// Module exports the module components
var Module = fx.Options(
	fx.Provide(controller.NewController),
)

type Handler struct {
	fx.In
	Controller controller.IProductController
	Logger     *logger.Logger
	DB         *postgres.DB
	Router     *routers.Router
}

func NewRoute(h Handler, m ...echo.MiddlewareFunc) Handler {
	h.Route(m...)
	return h
}

func (receiver *Handler) Route(m ...echo.MiddlewareFunc) {
	echoRoute := receiver.Router.Group("/v1/products", m...)
	echoRoute.POST("", receiver.CreateProduct)
	echoRoute.POST("/one-time-price", receiver.CreateOneTimePrice, receiver.Router.Authentication)
	// echoRoute.POST("/plan", receiver.CreatePlan)
	echoRoute.GET("", receiver.GetProducts)
}

func (receiver *Handler) CreateProduct(c echo.Context) error {
	req := new(product.CreateProductRequest)

	if err := c.Bind(req); err != nil {
		receiver.Logger.Error(err)
		return utilities.Response(c, &utilities.ResponseRequest{
			StatusCode: http.StatusBadRequest,
			Message:    utilities.BadRequest,
		})
	}

	tx := receiver.DB.Gorm.Begin()
	resp, err := receiver.Controller.CreateProduct(c.Request().Context(), req, tx)
	if err != nil {
		receiver.Logger.Error(err)
		tx.Rollback()

		parsedErr := utilities.ParseError(err)
		return utilities.Response(c, &utilities.ResponseRequest{
			StatusCode: parsedErr.StatusCode,
			Data:       parsedErr.Data,
			Message:    err.Error(),
		})
	}
	tx.Commit()

	return utilities.Response(c, &utilities.ResponseRequest{
		StatusCode: http.StatusCreated,
		Message:    utilities.Success,
		Data:       resp,
	})
}

func (receiver *Handler) CreateOneTimePrice(c echo.Context) error {
	req := new(product.CreateOneTimePriceRequest)

	data, ok := c.Request().Context().Value(jwt.InternalClaimData{}).(jwt.InternalClaimData)
	if !ok {
		return utilities.Response(c, &utilities.ResponseRequest{
			StatusCode: http.StatusUnauthorized,
			Message:    utilities.Authorization,
		})
	}
	req.ContextUserID = &data.UserID

	if err := c.Bind(req); err != nil {
		receiver.Logger.Error(err)
		return utilities.Response(c, &utilities.ResponseRequest{
			StatusCode: http.StatusBadRequest,
			Message:    utilities.BadRequest,
		})
	}

	tx := receiver.DB.Gorm.Begin()
	resp, err := receiver.Controller.CreateOneTimePrice(c.Request().Context(), req, tx)
	if err != nil {
		receiver.Logger.Error(err)
		tx.Rollback()

		parsedErr := utilities.ParseError(err)
		return utilities.Response(c, &utilities.ResponseRequest{
			StatusCode: parsedErr.StatusCode,
			Data:       parsedErr.Data,
			Message:    err.Error(),
		})
	}
	tx.Commit()

	return utilities.Response(c, &utilities.ResponseRequest{
		StatusCode: http.StatusCreated,
		Message:    utilities.Success,
		Data:       resp,
	})
}

/*
func (receiver *Handler) CreatePlan(c echo.Context) error {
	req := new(product.CreatePlanRequest)

	if err := c.Bind(req); err != nil {
		receiver.Logger.Error(err)
		return utilities.Response(c, &utilities.ResponseRequest{
			StatusCode: http.StatusBadRequest,
			Message:    utilities.BadRequest,
		})
	}

	tx := receiver.DB.Gorm.Begin()
	resp, err := receiver.Controller.CreatePlan(c.Request().Context(), req, tx)
	if err != nil {
		receiver.Logger.Error(err)
		tx.Rollback()

		parsedErr := utilities.ParseError(err)
		return utilities.Response(c, &utilities.ResponseRequest{
			StatusCode: parsedErr.StatusCode,
			Data:       parsedErr.Data,
			Message:    err.Error(),
		})
	}
	tx.Commit()

	return utilities.Response(c, &utilities.ResponseRequest{
		StatusCode: http.StatusCreated,
		Message:    utilities.Success,
		Data:       resp,
	})
}

*/

func (receiver *Handler) GetProducts(c echo.Context) error {
	resp, err := receiver.Controller.GetProducts(c.Request().Context(), nil)
	if err != nil {
		receiver.Logger.Error(err)

		parsedErr := utilities.ParseError(err)
		return utilities.Response(c, &utilities.ResponseRequest{
			StatusCode: parsedErr.StatusCode,
			Data:       parsedErr.Data,
			Message:    err.Error(),
		})
	}

	return utilities.Response(c, &utilities.ResponseRequest{
		StatusCode: http.StatusOK,
		Message:    utilities.Success,
		Data:       resp,
	})
}
