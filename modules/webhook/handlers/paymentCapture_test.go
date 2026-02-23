package handlers_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/demotask/backend/enum"
	"github.com/demotask/backend/models"
	"github.com/demotask/backend/modules/webhook/handlers"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/paypal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockOrderRepository
type MockOrderRepository struct{ mock.Mock }

func (m *MockOrderRepository) CreateOrder(ctx context.Context, order *models.Order, tx *gorm.DB) (*int, error) {
	return nil, nil
}
func (m *MockOrderRepository) FindByID(ctx context.Context, id int) (*models.Order, error) {
	return nil, nil
}
func (m *MockOrderRepository) FindByIDWithPayment(ctx context.Context, id int) (*models.Order, *models.Payment, error) {
	return nil, nil, nil
}
func (m *MockOrderRepository) FindByPaypalOrderID(ctx context.Context, req *models.Order) (*models.Order, error) {
	return nil, nil
}
func (m *MockOrderRepository) FindOrderID(ctx context.Context, req *models.Order) (*models.Order, error) {
	return nil, nil
}
func (m *MockOrderRepository) FindOrderPaypalID(ctx context.Context, paypalID string) (*models.Order, error) {
	args := m.Called(ctx, paypalID)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Order), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockOrderRepository) UpdateStatus(ctx context.Context, id int, status enum.OrderStatusInternal, tx *gorm.DB) error {
	args := m.Called(ctx, id, status, tx)
	return args.Error(0)
}
func (m *MockOrderRepository) UpdatePaypalOrderID(ctx context.Context, id int, paypalID string, tx *gorm.DB) error {
	return nil
}
func (m *MockOrderRepository) FindByUserID(ctx context.Context, userID int) ([]models.Order, error) {
	return nil, nil
}
func (m *MockOrderRepository) CreateOrderItems(ctx context.Context, items []models.OrderItem, tx *gorm.DB) error {
	return nil
}
func (m *MockOrderRepository) FindAllWithRelationsByUserID(ctx context.Context, userID int) ([]models.Order, error) {
	return nil, nil
}

func TestHandleCapture(t *testing.T) {
	db, sqlMock := setupTestDB(t)
	mockPayPal := new(MockPayPalClient)
	mockPayment := new(MockPaymentRepository)
	mockOrder := new(MockOrderRepository)
	log := logger.NewLogger()

	handler := handlers.NewPaymentCaptureHandler(handlers.PaymentCaptureHandler{
		DB:           db,
		Logger:       log,
		OrderRepo:    mockOrder,
		PaymentRepo:  mockPayment,
		PayPalClient: mockPayPal,
	})

	ctx := context.Background()
	payload := map[string]interface{}{
		"resource": map[string]interface{}{
			"supplementary_data": map[string]interface{}{
				"related_ids": map[string]interface{}{
					"order_id": "ORDER-123",
				},
			},
		},
	}

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(`^UPDATE "orders" SET "status"=\$1,"updated_at"=\$2 WHERE paypal_order_id = \$3`).
		WithArgs("PAID", sqlmock.AnyArg(), "ORDER-123").
		WillReturnResult(sqlmock.NewResult(1, 1))
	sqlMock.ExpectCommit()

	err := handler.HandleCapture(ctx, payload, db.Gorm)

	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestHandleVault(t *testing.T) {
	db, sqlMock := setupTestDB(t)
	mockPayPal := new(MockPayPalClient)
	mockPayment := new(MockPaymentRepository)
	mockOrder := new(MockOrderRepository)
	log := logger.NewLogger()

	handler := handlers.NewPaymentCaptureHandler(handlers.PaymentCaptureHandler{
		DB:           db,
		Logger:       log,
		OrderRepo:    mockOrder,
		PaymentRepo:  mockPayment,
		PayPalClient: mockPayPal,
	})

	ctx := context.Background()
	payload := map[string]interface{}{
		"resource": map[string]interface{}{
			"id": "VAULT-123",
			"metadata": map[string]interface{}{
				"order_id": "ORDER-123",
			},
		},
	}

	sqlMock.ExpectQuery(`^SELECT \* FROM "orders" WHERE paypal_order_id = \$1 AND "orders"\."deleted_at" IS NULL ORDER BY "orders"\."id" LIMIT \$2`).
		WithArgs("ORDER-123", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}).AddRow(1, 10))

	sqlMock.ExpectQuery(`^SELECT \* FROM "vault_tokens" WHERE paypal_vault_id = \$1 AND "vault_tokens"\."deleted_at" IS NULL ORDER BY "vault_tokens"\."id" LIMIT \$2`).
		WithArgs("VAULT-123", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(`^UPDATE "vault_tokens" SET "is_default"=\$1,"updated_at"=\$2 WHERE \(user_id = \$3 AND is_default = \$4\) AND "vault_tokens"\."deleted_at" IS NULL`).
		WithArgs(false, sqlmock.AnyArg(), 10, true).
		WillReturnResult(sqlmock.NewResult(1, 1))
	sqlMock.ExpectCommit()

	sqlMock.ExpectBegin()
	sqlMock.ExpectQuery(`^INSERT INTO "vault_tokens"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	sqlMock.ExpectCommit()

	err := handler.HandleVault(ctx, payload, db.Gorm)

	assert.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestHandleOrderApproved(t *testing.T) {
	db, _ := setupTestDB(t)
	mockPayPal := new(MockPayPalClient)
	mockPayment := new(MockPaymentRepository)
	mockOrder := new(MockOrderRepository)
	log := logger.NewLogger()

	handler := handlers.NewPaymentCaptureHandler(handlers.PaymentCaptureHandler{
		DB:           db,
		Logger:       log,
		OrderRepo:    mockOrder,
		PaymentRepo:  mockPayment,
		PayPalClient: mockPayPal,
	})

	ctx := context.Background()

	paypalOrderJSON := []byte(`{"id":"ORDER-123"}`)
	var rawResource map[string]interface{}
	json.Unmarshal(paypalOrderJSON, &rawResource)

	event := paypal.WebhookEvent{
		Resource: rawResource,
	}

	orderRecord := &models.Order{
		ID:     1,
		UserID: 10,
		Status: enum.OrderStatusInternalWaitingPayment,
	}
	mockOrder.On("FindOrderPaypalID", ctx, "ORDER-123").Return(orderRecord, nil)

	captureResp := &paypal.CaptureResponse{
		Status: "COMPLETED",
		PurchaseUnits: []paypal.PurchaseUnit{
			{
				Payments: &paypal.UnitPayments{
					Captures: []paypal.Capture{
						{
							ID: "CAP-123",
							Amount: paypal.Money{
								Value:        "10.00",
								CurrencyCode: "USD",
							},
						},
					},
				},
			},
		},
	}
	mockPayPal.On("CaptureOrder", "ORDER-123").Return(captureResp, nil)
	mockPayment.On("FindByCaptureID", ctx, "CAP-123").Return((*models.Payment)(nil), nil)

	mockPayment.On("Create", ctx, mock.AnythingOfType("*models.Payment"), db.Gorm).Return(nil, nil)
	mockOrder.On("UpdateStatus", ctx, 1, enum.OrderStatusInternalPaid, db.Gorm).Return(nil)

	err := handler.HandleOrderApproved(ctx, event, db.Gorm)

	assert.NoError(t, err)
	mockOrder.AssertExpectations(t)
	mockPayPal.AssertExpectations(t)
	mockPayment.AssertExpectations(t)
}
