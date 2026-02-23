package route

import (
	"net/http"

	"github.com/demotask/backend/modules/payment"
	"github.com/demotask/backend/modules/payment/controller"
	"github.com/demotask/backend/packages/jwt"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/postgres"
	"github.com/demotask/backend/routers"
	"github.com/demotask/backend/utilities"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

type Handler struct {
	fx.In
	Controller controller.IPaymentController
	Logger     *logger.Logger
	DB         *postgres.DB
	Router     *routers.Router
}

func NewRoute(h Handler, m ...echo.MiddlewareFunc) Handler {
	h.register(m...)
	return h
}

func (h *Handler) register(m ...echo.MiddlewareFunc) {
	group := h.Router.Group("/v1/payments", m...)

	// group.POST("/capture", h.Capture, h.Router.Authentication)
	group.POST("/pay-again", h.PayAgain, h.Router.Authentication)
	group.GET("/vault-tokens", h.GetVaultTokens, h.Router.Authentication)
}

func (h *Handler) PayAgain(c echo.Context) error {

	req := new(payment.PayAgainRequest)

	data, ok := c.Request().Context().
		Value(jwt.InternalClaimData{}).(jwt.InternalClaimData)
	if !ok {
		return utilities.Response(c, &utilities.ResponseRequest{
			StatusCode: http.StatusUnauthorized,
			Message:    utilities.Authorization,
		})
	}

	req.ContextUserID = &data.UserID

	if err := c.Bind(req); err != nil {
		return utilities.Response(c, &utilities.ResponseRequest{
			StatusCode: http.StatusBadRequest,
			Message:    utilities.BadRequest,
		})
	}

	tx := h.DB.Gorm.Begin()

	resp, err := h.Controller.PayAgain(c.Request().Context(), req, tx)
	if err != nil {
		tx.Rollback()
		parsed := utilities.ParseError(err)
		return utilities.Response(c, &utilities.ResponseRequest{
			StatusCode: parsed.StatusCode,
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

func (h *Handler) GetVaultTokens(c echo.Context) error {

	data, ok := c.Request().Context().
		Value(jwt.InternalClaimData{}).(jwt.InternalClaimData)
	if !ok {
		return utilities.Response(c, &utilities.ResponseRequest{
			StatusCode: http.StatusUnauthorized,
			Message:    utilities.Authorization,
		})
	}

	req := &payment.GetVaultTokensRequest{
		ContextUserID: &data.UserID,
	}

	resp, err := h.Controller.GetVaultTokens(c.Request().Context(), req)
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
		Data:       resp,
	})
}

// func (h *Handler) Capture(c echo.Context) error {

// 	req := new(payment.CapturePaymentRequest)

// 	if err := c.Bind(req); err != nil {
// 		h.Logger.Error(err)
// 		return utilities.Response(c, &utilities.ResponseRequest{
// 			StatusCode: http.StatusBadRequest,
// 			Message:    utilities.BadRequest,
// 		})
// 	}

// 	claim, ok := c.Request().Context().
// 		Value(jwt.InternalClaimData{}).(jwt.InternalClaimData)

// 	if !ok {
// 		return utilities.Response(c, &utilities.ResponseRequest{
// 			StatusCode: http.StatusUnauthorized,
// 			Message:    utilities.Authorization,
// 		})
// 	}

// 	req.ContextUserID = &claim.UserID

// 	tx := h.DB.Gorm.Begin()

// 	resp, err := h.Controller.Capture(c.Request().Context(), req, tx)

// 	if err != nil {
// 		tx.Rollback()
// 		h.Logger.Error(err)

// 		parsed := utilities.ParseError(err)

// 		return utilities.Response(c, &utilities.ResponseRequest{
// 			StatusCode: parsed.StatusCode,
// 			Data:       parsed.Data,
// 			Message:    err.Error(),
// 		})
// 	}

// 	tx.Commit()

// 	return utilities.Response(c, &utilities.ResponseRequest{
// 		StatusCode: http.StatusOK,
// 		Message:    utilities.Success,
// 		Data:       resp,
// 	})
// }
