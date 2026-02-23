-- +goose Up
CREATE TABLE product_one_time_prices (
    id           bigserial PRIMARY KEY,
    product_id   bigint NOT NULL REFERENCES products(id),
    price        decimal(10,2) NOT NULL,
    currency     varchar(3) NOT NULL DEFAULT 'USD',
    is_active    boolean NOT NULL DEFAULT true,

    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    deleted_at   timestamptz
);

CREATE INDEX idx_product_one_time_prices_product_id
ON product_one_time_prices(product_id);

CREATE INDEX idx_product_one_time_prices_deleted_at
ON product_one_time_prices(deleted_at);

-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
drop table product_one_time_prices;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
