package modules

import (
	billingPlan "github.com/demotask/backend/modules/billingPlan/route"
	oauth "github.com/demotask/backend/modules/oauth/route"
	order "github.com/demotask/backend/modules/order/route"
	payment "github.com/demotask/backend/modules/payment/route"

	// paymentmethod "github.com/demotask/backend/modules/paymentMethod/route"
	product "github.com/demotask/backend/modules/product/route"
	subscription "github.com/demotask/backend/modules/subscription/route"

	auth "github.com/demotask/backend/modules/auth/route"
	user "github.com/demotask/backend/modules/user/route"
	webhook "github.com/demotask/backend/modules/webhook/route"
	"go.uber.org/fx"
)

// AppRoute :
var AppRoute = fx.Options(
	fx.Invoke(auth.NewRoute),
	fx.Invoke(user.NewRoute),
	fx.Invoke(product.NewRoute),
	fx.Invoke(subscription.NewRoute),
	fx.Invoke(oauth.NewRoute),
	fx.Invoke(webhook.NewRoute),
	fx.Invoke(payment.NewRoute),
	fx.Invoke(order.NewRoute),
	fx.Invoke(billingPlan.NewRoute),
	// fx.Invoke(paymentmethod.NewRoute),
)
