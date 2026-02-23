-- +goose Up
CREATE TABLE billing_plans (
    id             BIGSERIAL PRIMARY KEY,
    product_id     BIGINT NOT NULL REFERENCES products(id),
    paypal_plan_id VARCHAR(255),

    name           VARCHAR(255) NOT NULL,
    description    TEXT,
    status         VARCHAR(50) NOT NULL,

    auto_bill_outstanding BOOLEAN,
    payment_failure_threshold INT,
    tax_percentage   DECIMAL(5,2),
    tax_inclusive    BOOLEAN,

    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX idx_billing_plans_product_id
    ON billing_plans(product_id);

CREATE INDEX idx_billing_plans_deleted_at
    ON billing_plans(deleted_at);

-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS billing_plans;

-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
