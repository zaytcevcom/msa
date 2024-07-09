-- +goose Up
-- +goose StatementBegin
CREATE TABLE accounts (
   id SERIAL PRIMARY KEY,
   user_id INT UNIQUE NOT NULL
);

CREATE TABLE payments (
   id SERIAL PRIMARY KEY,
   account_id INT NOT NULL,
   order_id INT,
   amount DECIMAL(10,2) NOT NULL
);

CREATE TABLE notifications (
   id SERIAL PRIMARY KEY,
   user_id INT NOT NULL,
   email VARCHAR(255) NOT NULL,
   text TEXT
);

CREATE TABLE orders (
   id SERIAL PRIMARY KEY,
   user_id INT NOT NULL,
   product_id INT NOT NULL,
   sum DECIMAL(10,2) NOT NULL,
   status INT NOT NULL,
   time INT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table accounts;
drop table payments;
drop table notifications;
drop table orders;
-- +goose StatementEnd
