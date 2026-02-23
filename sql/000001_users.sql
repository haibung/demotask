-- +goose Up
create table users (
     id            bigserial primary key,
     full_name     varchar(50),
     email         varchar(255) unique not null,
     password      varchar(255),

     updated_at    timestamptz default now(),
     created_at    timestamptz default now(),
     deleted_at    timestamptz default null
);
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
drop table users;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd