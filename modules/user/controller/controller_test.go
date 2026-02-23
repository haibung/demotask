package controller_test

import (
	"context"
	"testing"
	"time"

	"github.com/demotask/backend/models"
	"github.com/demotask/backend/modules/user"
	"github.com/demotask/backend/modules/user/controller"
	"github.com/demotask/backend/packages/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockUserRepository
type MockUserRepository struct{ mock.Mock }

func (m *MockUserRepository) Create(ctx context.Context, reqData *models.User, tx *gorm.DB) (*int, error) {
	args := m.Called(ctx, reqData, tx)
	if args.Get(0) != nil {
		id := args.Get(0).(int)
		return &id, args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockUserRepository) FindByID(ctx context.Context, reqData *models.User) (*models.User, error) {
	args := m.Called(ctx, reqData)
	if args.Get(0) != nil {
		return args.Get(0).(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockUserRepository) FindByEmail(ctx context.Context, reqData *models.User) (*models.User, error) {
	args := m.Called(ctx, reqData)
	if args.Get(0) != nil {
		return args.Get(0).(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestCreate(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	log := logger.NewLogger()

	c := controller.NewController(controller.UserController{
		UserRepository: mockUserRepo,
		Logger:         log,
	})

	ctx := context.Background()
	req := &user.CreateRequest{
		FullName: "Test User",
		Password: "Password123",
		Email:    "test@test.com",
	}

	mockUserRepo.On("FindByEmail", ctx, mock.AnythingOfType("*models.User")).Return(nil, gorm.ErrRecordNotFound)
	mockUserRepo.On("Create", ctx, mock.AnythingOfType("*models.User"), (*gorm.DB)(nil)).Return(1, nil)

	id, err := c.Create(ctx, req, nil)

	assert.NoError(t, err)
	assert.NotNil(t, id)
	assert.Equal(t, 1, *id)

	mockUserRepo.AssertExpectations(t)
}

func TestFindByEmail(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	log := logger.NewLogger()

	c := controller.NewController(controller.UserController{
		UserRepository: mockUserRepo,
		Logger:         log,
	})

	ctx := context.Background()
	req := &user.FindByEmailRequest{
		Email: "test@test.com",
	}

	now := time.Now()
	mockUser := &models.User{
		ID:        1,
		FullName:  "Test User",
		Email:     "test@test.com",
		Password:  "hashedpassword",
		CreatedAt: now,
	}

	mockUserRepo.On("FindByEmail", ctx, mock.AnythingOfType("*models.User")).Return(mockUser, nil)

	resp, err := c.FindByEmail(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, resp.ID)
	assert.Equal(t, "test@test.com", resp.Email)
	assert.Equal(t, "Test User", resp.FullName)

	mockUserRepo.AssertExpectations(t)
}
