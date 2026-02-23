-- +goose Up
CREATE TABLE subscriptions (
    id                      BIGSERIAL PRIMARY KEY,
    user_id                 BIGINT NOT NULL REFERENCES users(id),
    order_id                BIGINT NOT NULL REFERENCES orders(id),
    billing_plan_id         BIGINT NOT NULL REFERENCES billing_plans(id),
    paypal_subscription_id  VARCHAR(255) NOT NULL UNIQUE,
    status                  VARCHAR(50) NOT NULL,
    start_time              TIMESTAMPTZ,
    vault_token_id          BIGINT NULL REFERENCES vault_tokens(id),
    quantity                INT NOT NULL,
    next_billing_time       TIMESTAMPTZ,
    failed_payments_count   INT DEFAULT 0,
    last_payment_time       TIMESTAMPTZ,
    snapshot                JSONB,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at              TIMESTAMPTZ
);

CREATE INDEX idx_subscriptions_user_id
    ON subscriptions(user_id);

CREATE INDEX idx_subscriptions_deleted_at
    ON subscriptions(deleted_at);

-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
drop table subscriptions;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
