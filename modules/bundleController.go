package modules

import (
	auth "github.com/demotask/backend/modules/auth/controller"
	billingPlan "github.com/demotask/backend/modules/billingPlan/controller"
	oauth "github.com/demotask/backend/modules/oauth/controller"
	order "github.com/demotask/backend/modules/order/controller"
	payment "github.com/demotask/backend/modules/payment/controller"

	// paymentmethod "github.com/demotask/backend/modules/paymentMethod/controller"
	product "github.com/demotask/backend/modules/product/controller"
	subscription "github.com/demotask/backend/modules/subscription/controller"
	user "github.com/demotask/backend/modules/user/controller"

	webhook "github.com/demotask/backend/modules/webhook/controller"
	"github.com/demotask/backend/modules/webhook/handlers"
	"github.com/demotask/backend/packages/paypal"
	"go.uber.org/fx"
)

// AppController :
var AppController = fx.Options(
	fx.Provide(auth.NewController),
	fx.Provide(product.NewController),
	fx.Provide(subscription.NewController),
	fx.Provide(oauth.NewController),
	fx.Provide(webhook.NewController),
	fx.Provide(payment.NewController),
	fx.Provide(paypal.NewPayPalClient),
	fx.Provide(handlers.NewPaymentCaptureHandler),
	// fx.Provide(handlers.NewSubscriptionRenewalHandler),
	fx.Provide(handlers.NewSubscriptionHandler),
	fx.Provide(user.NewController),
	fx.Provide(billingPlan.NewController),
	fx.Provide(order.NewController),
	// fx.Provide(paymentmethod.NewController),
)
