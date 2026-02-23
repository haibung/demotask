-- +goose Up
CREATE TABLE vault_tokens (
    id               BIGSERIAL PRIMARY KEY,
    user_id          BIGINT NOT NULL REFERENCES users(id),
    paypal_vault_id  VARCHAR(255) NOT NULL UNIQUE,
    email            VARCHAR(255),
    is_default       BOOLEAN NOT NULL DEFAULT false,

    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at       TIMESTAMPTZ
);

CREATE INDEX idx_vault_tokens_user_id
    ON vault_tokens(user_id);

CREATE INDEX idx_vault_tokens_deleted_at
    ON vault_tokens(deleted_at);

-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
drop table vault_tokens;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
