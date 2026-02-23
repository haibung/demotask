package routers

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/labstack/gommon/log"
	"github.com/sirupsen/logrus"

	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/utilities"

	userRepository "github.com/demotask/backend/modules/user/repository"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Router struct {
	*echo.Echo
	logger         *logger.Logger
	userRepository userRepository.IUserInterface
}

//var RouteLog *logger.Logger

func NewRouter(userRepository userRepository.IUserInterface, logger *logger.Logger) *Router {

	var (
		e     = echo.New()
		route = Router{
			e, logger, userRepository,
		}
	)

	//Set middleware
	e.Pre(middleware.RemoveTrailingSlash())
	e.Static("/static", "static")
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			route.logger.Request(c.Request().Header.Get("X-Request-ID"))
			RequestLog := route.logger.WithFields(logrus.Fields{
				"path":       c.Request().URL.Path,
				"method":     c.Request().Method,
				"version":    c.Request().Header.Get("Version"),
				"queryParam": c.QueryParams(),
			})

			//If content-type not application/json will skip
			if c.Request().Header.Get("Content-Type") != echo.MIMEApplicationJSON {
				RequestLog.Info()
				return next(c)
			}

			//Read reqData from body
			reqData, err := ioutil.ReadAll(c.Request().Body)
			if err != nil {
				route.logger.Error(err.Error())
				return utilities.Response(c, &utilities.ResponseRequest{
					StatusCode: http.StatusInternalServerError,
					Message:    err.Error(),
				})
			}

			//If request method POST | PUT | DELETE will print reqData
			if c.Request().Method == http.MethodPost ||
				c.Request().Method == http.MethodPut ||
				c.Request().Method == http.MethodDelete {
				RequestLog.Info(string(reqData))
			}

			c.Request().Body = ioutil.NopCloser(bytes.NewReader(reqData))
			return next(c)
		}
	})
	e.Use(middleware.BodyLimit("10M"))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			"Authorization",
			"Version",
		},
	}))
	e.Use(middleware.RateLimiterWithConfig(rateLimitConfig()))
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		Skipper:   middleware.DefaultSkipper,
		StackSize: 1 << 10,
		LogLevel:  log.ERROR,
	}))

	return &route
}

func rateLimitConfig() middleware.RateLimiterConfig {
	return middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{Rate: 5, Burst: 15, ExpiresIn: 3 * time.Minute},
		),
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			id := ctx.RealIP()
			return id, nil
		},
		ErrorHandler: func(context echo.Context, err error) error {
			return utilities.Response(context, &utilities.ResponseRequest{
				StatusCode: http.StatusTooManyRequests,
				Message:    utilities.Forbidden,
			})
		},
		DenyHandler: func(context echo.Context, identifier string, err error) error {
			return utilities.Response(context, &utilities.ResponseRequest{
				StatusCode: http.StatusTooManyRequests,
				Message:    utilities.Forbidden,
			})
		},
	}
}
