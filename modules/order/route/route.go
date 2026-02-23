package route

import (
	"net/http"
	"strconv"

	"github.com/demotask/backend/modules/order"
	"github.com/demotask/backend/modules/order/controller"
	"github.com/demotask/backend/packages/jwt"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/postgres"
	"github.com/demotask/backend/routers"
	"github.com/demotask/backend/utilities"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(controller.NewController),
)

type Handler struct {
	fx.In

	Controller controller.IOrderController
	Logger     *logger.Logger
	DB         *postgres.DB
	Router     *routers.Router
}

func NewRoute(h Handler, m ...echo.MiddlewareFunc) Handler {
	h.Route(m...)
	return h
}

func (receiver *Handler) Route(m ...echo.MiddlewareFunc) {
	echoRoute := receiver.Router.Group("/v1/orders", m...)

	echoRoute.POST("", receiver.CreateOrder, receiver.Router.Authentication)
	echoRoute.POST("/:id/pay", receiver.PayOrder, receiver.Router.Authentication)
	echoRoute.GET("/:id", receiver.GetOrder, receiver.Router.Authentication)
	echoRoute.GET("", receiver.GetMyOrders, receiver.Router.Authentication)
}

func (receiver *Handler) CreateOrder(c echo.Context) error {
	req := new(order.CreateOrderRequest)

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
	resp, err := receiver.Controller.CreateOrder(c.Request().Context(), req, tx)
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

func (h *Handler) PayOrder(c echo.Context) error {
	req := new(order.PayOrderRequest)

	data, ok := c.Request().Context().Value(jwt.InternalClaimData{}).(jwt.InternalClaimData)
	if !ok {
		return utilities.Response(c, &utilities.ResponseRequest{
			StatusCode: http.StatusUnauthorized,
			Message:    utilities.Authorization,
		})
	}
	req.ContextUserID = &data.UserID

	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil || orderID <= 0 {
		return utilities.Response(c, &utilities.ResponseRequest{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid order id",
		})
	}
	req.OrderID = orderID

	if err := c.Bind(req); err != nil {
		h.Logger.Error(err)
		return utilities.Response(c, &utilities.ResponseRequest{
			StatusCode: http.StatusBadRequest,
			Message:    utilities.BadRequest,
		})
	}

	tx := h.DB.Gorm.Begin()
	resp, err := h.Controller.PayOrder(c.Request().Context(), req, tx)
	if err != nil {
		tx.Rollback()
		parsed := utilities.ParseError(err)
		return utilities.Response(c, &utilities.ResponseRequest{
			StatusCode: parsed.StatusCode,
			Data:       parsed.Data,
			Message:    err.Error(),
		})
	}
	tx.Commit()

	return utilities.Response(c, &utilities.ResponseRequest{
		StatusCode: http.StatusOK,
		Message:    utilities.Success,
		Data:       resp,
	})
}
func (h *Handler) GetOrder(c echo.Context) error {
	data, ok := c.Request().Context().
		Value(jwt.InternalClaimData{}).(jwt.InternalClaimData)

	if !ok {
		return utilities.Response(c, &utilities.ResponseRequest{
			StatusCode: http.StatusUnauthorized,
			Message:    utilities.Authorization,
		})
	}

	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil || orderID <= 0 {
		return utilities.Response(c, &utilities.ResponseRequest{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid order id",
		})
	}

	resp, err := h.Controller.GetOrderByID(
		c.Request().Context(),
		orderID,
		data.UserID,
	)
	if err != nil {
		parsed := utilities.ParseError(err)
		return utilities.Response(c, &utilities.ResponseRequest{
			StatusCode: parsed.StatusCode,
			Data:       parsed.Data,
			Message:    err.Error(),
		})
	}

	return utilities.Response(c, &utilities.ResponseRequest{
		StatusCode: http.StatusOK,
		Message:    utilities.Success,
		Data:       resp,
	})
}

func (h *Handler) GetMyOrders(c echo.Context) error {
	data, ok := c.Request().Context().Value(jwt.InternalClaimData{}).(jwt.InternalClaimData)
	if !ok {
		return utilities.Response(c, &utilities.ResponseRequest{
			StatusCode: http.StatusUnauthorized,
			Message:    utilities.Authorization,
		})
	}

	resp, err := h.Controller.GetAllOrders(c.Request().Context(), data.UserID)
	if err != nil {
		parsed := utilities.ParseError(err)
		return utilities.Response(c, &utilities.ResponseRequest{
			StatusCode: parsed.StatusCode,
			Data:       parsed.Data,
			Message:    err.Error(),
		})
	}

	return utilities.Response(c, &utilities.ResponseRequest{
		StatusCode: http.StatusOK,
		Message:    utilities.Success,
		Data:       resp,
	})
}
