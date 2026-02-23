package controller_test

import (
	"context"
	"testing"
	"time"

	"github.com/demotask/backend/config"
	"github.com/demotask/backend/modules/auth"
	"github.com/demotask/backend/modules/auth/controller"
	"github.com/demotask/backend/modules/user"
	"github.com/demotask/backend/packages/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// MockUserController
type MockUserController struct{ mock.Mock }

func (m *MockUserController) Create(ctx context.Context, reqData *user.CreateRequest, tx *gorm.DB) (*int, error) {
	args := m.Called(ctx, reqData, tx)
	if args.Get(0) != nil {
		id := args.Get(0).(int)
		return &id, args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockUserController) FindByEmail(ctx context.Context, reqData *user.FindByEmailRequest) (*user.FindByEmailResponse, error) {
	args := m.Called(ctx, reqData)
	if args.Get(0) != nil {
		return args.Get(0).(*user.FindByEmailResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestLogin(t *testing.T) {
	mockUserController := new(MockUserController)
	log := logger.NewLogger()

	c := controller.NewController(controller.AuthController{
		Config:         &config.Config{},
		UserController: mockUserController,
		Logger:         log,
	})

	ctx := context.Background()
	req := &auth.LoginRequest{
		Email:    "test@test.com",
		Password: "Password123",
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Password123"), bcrypt.MinCost)

	mockUserResp := &user.FindByEmailResponse{
		ID:        1,
		FullName:  "Test User",
		Email:     "test@test.com",
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
	}

	mockUserController.On("FindByEmail", ctx, mock.AnythingOfType("*user.FindByEmailRequest")).Return(mockUserResp, nil)

	resp, err := c.Login(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)

	mockUserController.AssertExpectations(t)
}

func TestRegister(t *testing.T) {
	mockUserController := new(MockUserController)
	log := logger.NewLogger()

	c := controller.NewController(controller.AuthController{
		Config:         &config.Config{},
		UserController: mockUserController,
		Logger:         log,
	})

	ctx := context.Background()
	req := &auth.RegisterRequest{
		FullName: "Test User",
		Email:    "test@test.com",
		Password: "Password123",
	}

	mockUserController.On("Create", ctx, mock.AnythingOfType("*user.CreateRequest"), (*gorm.DB)(nil)).Return(1, nil)

	err := c.Register(ctx, req, nil)

	assert.NoError(t, err)

	mockUserController.AssertExpectations(t)
}
