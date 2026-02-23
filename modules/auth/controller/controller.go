package controller

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/demotask/backend/config"
	"github.com/demotask/backend/modules/auth"
	userDto "github.com/demotask/backend/modules/user"
	userController "github.com/demotask/backend/modules/user/controller"
	_jwt "github.com/demotask/backend/packages/jwt"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/utilities"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt"
	"go.uber.org/fx"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type (
	IAuthController interface {
		Login(ctx context.Context, reqData *auth.LoginRequest) (*auth.LoginResponse, error)
		Register(ctx context.Context, reqData *auth.RegisterRequest, tx *gorm.DB) error
	}

	AuthController struct {
		fx.In
		Config         *config.Config
		Logger         *logger.Logger
		UserController userController.IUserController
	}
)

// NewController :
func NewController(authController AuthController) IAuthController {
	return &authController
}

// Login :
func (receiver *AuthController) Login(ctx context.Context, reqData *auth.LoginRequest) (*auth.LoginResponse, error) {
	// Validate request data
	messages, err := reqData.Validate()
	if err != nil {
		receiver.Logger.Error(err)
		return nil, utilities.ErrorRequest(err, http.StatusBadRequest, messages)
	}

	fetchUser, err := receiver.UserController.FindByEmail(ctx, &userDto.FindByEmailRequest{
		Email: reqData.Email,
	})
	if err != nil {
		receiver.Logger.Error(err)

		if utilities.ParseError(err).StatusCode == http.StatusNotFound {
			return nil, utilities.ErrorRequest(errors.New(utilities.InvalidAccessLogin), http.StatusForbidden)
		}

		return nil, err
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(fetchUser.Password), []byte(reqData.Password)); err != nil {
		receiver.Logger.Error(err)
		return nil, utilities.ErrorRequest(errors.New(utilities.InvalidAccessLogin), http.StatusForbidden)
	}

	// Generate uuid for user jwt
	generateUUID, err := uuid.NewV4()
	if err != nil {
		receiver.Logger.Error(err)
		return nil, utilities.ErrorRequest(err, http.StatusInternalServerError)
	}

	// Generate access and refresh token
	accessToken, err := _jwt.GenerateToken(_jwt.Claim{
		Data: _jwt.ClaimData{
			UserID: fetchUser.ID,
			UUID:   generateUUID.String(),
		},
		StandardClaims: jwt.StandardClaims{
			Audience:  "", // Web | Mobile = Get from context header
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(receiver.Config.Auth.ExpireAccessTokenDuration).Unix(),
		},
	}, receiver.Config.Auth.Secret)
	if err != nil {
		receiver.Logger.Error(err)
		return nil, utilities.ErrorRequest(err, http.StatusInternalServerError)
	}

	refreshToken, err := _jwt.GenerateToken(_jwt.Claim{
		Data: _jwt.ClaimData{
			UserID: fetchUser.ID,
			UUID:   generateUUID.String(),
		},
		StandardClaims: jwt.StandardClaims{
			Audience:  "", // Web | Mobile = Get from context header
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(receiver.Config.Auth.ExpireRefreshTokenDuration).Unix(),
		},
	}, receiver.Config.Auth.SecretClaim)
	if err != nil {
		receiver.Logger.Error(err)
		return nil, utilities.ErrorRequest(err, http.StatusInternalServerError)
	}

	return &auth.LoginResponse{
		AccessToken:  *accessToken,
		RefreshToken: *refreshToken,
	}, nil
}

// Register
func (l *AuthController) Register(ctx context.Context, reqData *auth.RegisterRequest, tx *gorm.DB) error {
	// Validate request data
	messages, err := reqData.Validate()
	if err != nil {
		l.Logger.Error(err)
		return utilities.ErrorRequest(err, http.StatusBadRequest, messages)
	}
	if _, err := l.UserController.Create(ctx, &userDto.CreateRequest{
		FullName: reqData.FullName,
		Password: reqData.Password,
		Email:    reqData.Email,
	}, tx); err != nil {
		l.Logger.Error(err)
		return utilities.ErrorRequest(err, http.StatusInternalServerError)
	}

	return nil
}
