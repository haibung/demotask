package route

import (
	"net/http"

	subscription "github.com/demotask/backend/modules/billingPlan"
	"github.com/demotask/backend/modules/billingPlan/controller"
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
	Controller controller.IBillingPlanController
	Logger     *logger.Logger
	DB         *postgres.DB
	Router     *routers.Router
}

func NewRoute(h Handler, m ...echo.MiddlewareFunc) Handler {
	h.Route(m...)
	return h
}

func (receiver *Handler) Route(m ...echo.MiddlewareFunc) {
	echoRoute := receiver.Router.Group("/v1/billing-plans", m...)
	echoRoute.POST("", receiver.CreateBillingPlan, receiver.Router.Authentication)
}

func (receiver *Handler) CreateBillingPlan(c echo.Context) error {
	req := new(subscription.CreateBillingPlanRequest)

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
	resp, err := receiver.Controller.CreateBillingPlan(c.Request().Context(), req, tx)
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
