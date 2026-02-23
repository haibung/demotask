package modules

import (
	billingPlan "github.com/demotask/backend/modules/billingPlan/repository"
	oauth "github.com/demotask/backend/modules/oauth/repository"
	payment "github.com/demotask/backend/modules/payment/repository"
	user "github.com/demotask/backend/modules/user/repository"

	order "github.com/demotask/backend/modules/order/repository"
	// paymentmethod "github.com/demotask/backend/modules/paymentMethod/repository"
	product "github.com/demotask/backend/modules/product/repository"
	subscription "github.com/demotask/backend/modules/subscription/repository"

	webhook "github.com/demotask/backend/modules/webhook/repository"
	"go.uber.org/fx"
)

// AppRepository :
var AppRepository = fx.Options(
	fx.Provide(product.NewRepository),
	fx.Provide(subscription.NewRepository),
	fx.Provide(oauth.NewRepository),
	fx.Provide(webhook.NewRepository),
	fx.Provide(payment.NewRepository),
	fx.Provide(user.NewRepository),
	fx.Provide(billingPlan.NewRepository),
	fx.Provide(order.NewRepository),
	// fx.Provide(paymentmethod.NewRepository),
)
