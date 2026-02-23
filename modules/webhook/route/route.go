package route

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/demotask/backend/modules/webhook/controller"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/postgres"
	"github.com/demotask/backend/routers"
	"github.com/demotask/backend/utilities"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

type Handler struct {
	fx.In
	Controller controller.IWebhookController
	Logger     *logger.Logger
	DB         *postgres.DB
	Router     *routers.Router
}

func NewRoute(h Handler) Handler {
	h.Route()
	return h
}

func (h *Handler) Route() {
	group := h.Router.Group("/v1/webhooks")
	group.POST("/paypal", h.HandlePayPalWebhook)
}

func (h *Handler) HandlePayPalWebhook(c echo.Context) error {
	reqBody, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return utilities.Response(c, &utilities.ResponseRequest{
			StatusCode: http.StatusBadRequest,
			Message:    "Failed to read request body",
		})
	}

	var body map[string]interface{}
	if err := json.Unmarshal(reqBody, &body); err != nil {
		return utilities.Response(c, &utilities.ResponseRequest{
			StatusCode: http.StatusBadRequest,
			Message:    utilities.BadRequest,
		})
	}

	headers := map[string]string{
		"paypal-auth-algo":         c.Request().Header.Get("Paypal-Auth-Algo"),
		"paypal-cert-url":          c.Request().Header.Get("Paypal-Cert-Url"),
		"paypal-transmission-id":   c.Request().Header.Get("Paypal-Transmission-Id"),
		"paypal-transmission-sig":  c.Request().Header.Get("Paypal-Transmission-Sig"),
		"paypal-transmission-time": c.Request().Header.Get("Paypal-Transmission-Time"),
		"paypal-webhook-id":        c.Request().Header.Get("Paypal-Webhook-Id"),
	}

	tx := h.DB.Gorm.Begin()
	err = h.Controller.HandlePayPalWebhook(c.Request().Context(), headers, reqBody, body, tx)

	if err != nil {
		tx.Rollback()
		parsed := utilities.ParseError(err)
		return utilities.Response(c, &utilities.ResponseRequest{
			StatusCode: parsed.StatusCode,
			Message:    err.Error(),
			Data:       parsed.Data,
		})
	}

	tx.Commit()
	return c.NoContent(http.StatusOK)
}

//

// func (h *Handler) HandlePayPalWebhook(c echo.Context) error {

// 	reqBody, err := io.ReadAll(c.Request().Body)
// 	if err != nil {
// 		return utilities.Response(c, &utilities.ResponseRequest{
// 			StatusCode: http.StatusBadRequest,
// 			Message:    "invalid body",
// 		})
// 	}

// 	headers := map[string]string{
// 		"paypal-auth-algo":         c.Request().Header.Get("Paypal-Auth-Algo"),
// 		"paypal-cert-url":          c.Request().Header.Get("Paypal-Cert-Url"),
// 		"paypal-transmission-id":   c.Request().Header.Get("Paypal-Transmission-Id"),
// 		"paypal-transmission-sig":  c.Request().Header.Get("Paypal-Transmission-Sig"),
// 		"paypal-transmission-time": c.Request().Header.Get("Paypal-Transmission-Time"),
// 	}

// 	tx := h.DB.Gorm.Begin()

// 	err = h.Controller.HandlePayPalWebhook(
// 		c.Request().Context(),
// 		headers,
// 		reqBody,
// 		tx,
// 	)

// 	if err != nil {
// 		tx.Rollback()
// 		parsed := utilities.ParseError(err)
// 		return utilities.Response(c, &utilities.ResponseRequest{
// 			StatusCode: parsed.StatusCode,
// 			Message:    parsed.Message,
// 			Data:       parsed.Data,
// 		})
// 	}

// 	tx.Commit()
// 	return c.NoContent(http.StatusOK)
// }
