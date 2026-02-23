package routers

import (
	"context"
	"net/http"
	"strings"

	"github.com/demotask/backend/config"
	"github.com/demotask/backend/models"
	"github.com/demotask/backend/packages/jwt"
	"github.com/demotask/backend/utilities"
	"github.com/labstack/echo/v4"
)

func (receiver *Router) Authentication(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authorizationHeader := c.Request().Header.Get("Authorization")
		authorization := strings.Split(authorizationHeader, " ")
		if len(authorization) > 1 {
			result, err := jwt.ParseClaim(authorization[1], config.Get().Auth.Secret)
			if err != nil {
				receiver.Logger.Error(err.Error())
				return utilities.Response(c, &utilities.ResponseRequest{
					StatusCode: http.StatusUnauthorized,
					Message:    utilities.Authorization,
				})
			}

			var (
				ctx = c.Request().Context()
			)

			// Check exist user
			fetchUser, err := receiver.userRepository.FindByID(ctx, &models.User{
				ID: result.Data.UserID,
			})
			if err != nil {
				receiver.logger.Error(err.Error())
				return utilities.Response(c, &utilities.ResponseRequest{
					StatusCode: http.StatusUnauthorized,
					Message:    utilities.Authorization,
				})
			}

			ctx = context.WithValue(ctx, jwt.InternalClaimData{}, jwt.InternalClaimData{
				UserID: fetchUser.ID,
			})

			c.SetRequest(c.Request().WithContext(ctx))
		} else {
			return utilities.Response(c, &utilities.ResponseRequest{
				StatusCode: http.StatusUnauthorized,
				Message:    utilities.Authorization,
			})
		}
		return next(c)
	}
}
