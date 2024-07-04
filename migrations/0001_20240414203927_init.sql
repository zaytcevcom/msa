-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
   id SERIAL PRIMARY KEY,
   username VARCHAR(255) UNIQUE NOT NULL,
   password_hash VARCHAR(64) NOT NULL,
   first_name VARCHAR(255) NOT NULL,
   last_name VARCHAR(255) NOT NULL,
   email VARCHAR(255) NOT NULL,
   phone VARCHAR(20) NOT NULL
);
CREATE UNIQUE INDEX users_username_idx ON users (username);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table users
-- +goose StatementEnd
