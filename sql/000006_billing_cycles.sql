-- +goose Up
CREATE TABLE billing_cycles (
    id               BIGSERIAL PRIMARY KEY,
    billing_plan_id  BIGINT NOT NULL REFERENCES billing_plans(id),
    interval_unit    VARCHAR(10) NOT NULL,  -- DAY, WEEK, MONTH, YEAR
    interval_count   INT NOT NULL,           -- 1, 3, 6, 12
    tenure_type      VARCHAR(10) NOT NULL,   -- TRIAL, REGULAR
    sequence         INT NOT NULL,
    total_cycles     INT NOT NULL,            -- 0 = unlimited
    price_value      DECIMAL(10,2) NOT NULL,
    currency         VARCHAR(3) NOT NULL,

    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at       TIMESTAMPTZ
);

CREATE INDEX idx_billing_cycles_billing_plan_id
    ON billing_cycles(billing_plan_id);

CREATE INDEX idx_billing_cycles_deleted_at
    ON billing_cycles(deleted_at);
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
drop table billing_cycles;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
