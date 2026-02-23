-- +goose Up
CREATE TABLE products (
    id                bigserial PRIMARY KEY,
    paypal_product_id varchar(255),
    name              varchar(255) NOT NULL,
    description       text,
    business_type     varchar(20) NOT NULL DEFAULT 'SERVICE',
    category          varchar(100),
    image_url         text,
    home_url          text,

    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now(),
    deleted_at        timestamptz
);

CREATE INDEX idx_products_deleted_at ON products(deleted_at);

-- +goose Down
DROP TABLE products;