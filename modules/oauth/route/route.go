package route

import (
	"net/http"

	"github.com/demotask/backend/modules/oauth/controller"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/paypal"
	"github.com/demotask/backend/packages/postgres"
	"github.com/demotask/backend/routers"
	"github.com/demotask/backend/utilities"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

type Handler struct {
	fx.In
	Controller   controller.IOAuthController
	Logger       *logger.Logger
	DB           *postgres.DB
	Router       *routers.Router
	PayPalClient paypal.IPayPalClient
}

func NewRoute(h Handler, m ...echo.MiddlewareFunc) Handler {
	h.Route(m...)
	return h
}

func (receiver *Handler) Route(m ...echo.MiddlewareFunc) {
	publicRoute := receiver.Router.Group("/v1/oauth")
	// publicRoute.GET("/callback", receiver.HandlePayPalCallback)
	// publicRoute.GET("/paypal/callback", receiver.HandlePayPalCallback)
	publicRoute.POST("/login", receiver.GetClientCredentialsToken)
}

// GetClientCredentialsToken generates a client credentials token
func (receiver *Handler) GetClientCredentialsToken(c echo.Context) error {
	resp, err := receiver.Controller.GetClientToken(c.Request().Context())
	if err != nil {
		receiver.Logger.Error(err)
		return utilities.Response(c, &utilities.ResponseRequest{
			StatusCode: utilities.ParseError(err).StatusCode,
			Data:       utilities.ParseError(err).Data,
			Message:    err.Error(),
		})
	}

	return utilities.Response(c, &utilities.ResponseRequest{
		StatusCode: http.StatusOK,
		Message:    utilities.Success,
		Data:       resp,
	})
}

// HandlePayPalCallback handles the OAuth callback from PayPal
// func (receiver *Handler) HandlePayPalCallback(c echo.Context) error {
// 	code := c.QueryParam("code")
// 	state := c.QueryParam("state")

// 	var reqData = &oauth.OAuthCallbackRequest{
// 		Provider: "paypal",
// 		Code:     code,
// 		State:    state,
// 	}

// 	tx := receiver.DB.Gorm.Begin()
// 	resp, err := receiver.Controller.HandleCallback(c.Request().Context(), reqData, tx)
// 	if err != nil {
// 		receiver.Logger.Error(err)
// 		defer func() {
// 			tx.Rollback()
// 		}()

// 		return utilities.Response(c, &utilities.ResponseRequest{
// 			StatusCode: utilities.ParseError(err).StatusCode,
// 			Data:       utilities.ParseError(err).Data,
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
