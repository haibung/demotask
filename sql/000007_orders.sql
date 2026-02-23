-- +goose Up
CREATE TABLE orders (
    id               BIGSERIAL PRIMARY KEY,
    user_id          BIGINT NOT NULL REFERENCES users(id),
    paypal_order_id  VARCHAR(255),
    status           VARCHAR(50) NOT NULL,
    order_type       VARCHAR(50),

    snapshot         JSONB,

    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at       TIMESTAMPTZ
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_deleted_at ON orders(deleted_at);
CREATE INDEX idx_orders_paypal_order_id ON orders(paypal_order_id);

-- +goose Down
DROP TABLE orders;
