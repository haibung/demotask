package route

import (
	"net/http"

	"github.com/demotask/backend/modules/subscription"
	"github.com/demotask/backend/modules/subscription/controller"
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
	Controller controller.ISubscriptionController
	Logger     *logger.Logger
	DB         *postgres.DB
	Router     *routers.Router
}

func NewRoute(h Handler, m ...echo.MiddlewareFunc) Handler {
	h.Route(m...)
	return h
}

func (receiver *Handler) Route(m ...echo.MiddlewareFunc) {
	echoRoute := receiver.Router.Group("/v1/subscriptions", m...)
	echoRoute.POST("", receiver.CreateSubscription, receiver.Router.Authentication)
	// echoRoute.GET("/:id", receiver.GetSubscription, receiver.Router.Authentication)
	echoRoute.GET("", receiver.GetMySubscriptions, receiver.Router.Authentication)
}

func (receiver *Handler) CreateSubscription(c echo.Context) error {
	req := new(subscription.CreateSubscriptionRequest)

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
	resp, err := receiver.Controller.CreateSubscription(c.Request().Context(), req, tx)
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

// TODO : PAYMENT BELUM BISA DIAKSES
// TODO : GET PRODUCT BELUM DI BUAT

func (h *Handler) GetMySubscriptions(c echo.Context) error {

	claim, ok := c.Request().Context().
		Value(jwt.InternalClaimData{}).(jwt.InternalClaimData)

	if !ok {
		return utilities.Response(c, &utilities.ResponseRequest{
			StatusCode: http.StatusUnauthorized,
			Message:    utilities.Authorization,
		})
	}

	subs, err := h.Controller.GetMySubscriptions(
		c.Request().Context(),
		claim.UserID,
	)

	if err != nil {
		parsed := utilities.ParseError(err)
		return utilities.Response(c, &utilities.ResponseRequest{
			StatusCode: parsed.StatusCode,
			Message:    err.Error(),
		})
	}

	return utilities.Response(c, &utilities.ResponseRequest{
		StatusCode: http.StatusOK,
		Message:    utilities.Success,
		Data:       subs,
	})
}

// func (receiver *Handler) GetSubscription(c echo.Conte
// func (h *Handler) GetSubscription(c echo.Context) error {
// 	req := new(subscription.GetSubscriptionRequest)

// 	data, ok := c.Request().Context().Value(jwt.InternalClaimData{}).(jwt.InternalClaimData)
// 	if !ok {
// 		return utilities.Response(c, &utilities.ResponseRequest{
// 			StatusCode: http.StatusUnauthorized,
// 			Message:    utilities.Authorization,
// 		})
// 	}
// 	req.ContextUserID = &data.UserID

// 	idParam := c.Param("id")

// 	id, err := strconv.Atoi(idParam)
// 	if err != nil {
// 		return utilities.Response(c, &utilities.ResponseRequest{
// 			StatusCode: http.StatusBadRequest,
// 			Message:    "Invalid ID",
// 		})
// 	}

// 	sub, err := h.Controller.GetSubscription(c.Request().Context(), id)

// 	if err != nil {
// 		parsed := utilities.ParseError(err)
// 		return utilities.Response(c, &utilities.ResponseRequest{
// 			StatusCode: parsed.StatusCode,
// 			Message:    err.Error(),
// 		})
// 	}

// 	return utilities.Response(c, &utilities.ResponseRequest{
// 		StatusCode: http.StatusOK,
// 		Message:    utilities.Success,
// 		Data:       sub,
// 	})
// }
