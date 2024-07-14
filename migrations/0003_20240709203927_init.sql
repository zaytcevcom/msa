-- +goose Up
-- +goose StatementBegin
CREATE TABLE products (
   id SERIAL PRIMARY KEY,
   name VARCHAR(255) UNIQUE NOT NULL,
   count INT NOT NULL
);

CREATE TABLE product_reserve (
   id SERIAL PRIMARY KEY,
   order_id INT,
   product_id INT NOT NULL,
   count INT NOT NULL
);

CREATE TABLE employees (
   id SERIAL PRIMARY KEY,
   name VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE employee_reserve (
   id SERIAL PRIMARY KEY,
   order_id INT NOT NULL,
   employee_id INT NOT NULL,
   time INT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table products;
drop table product_reserve;
drop table employees;
drop table employee_reserve;
-- +goose StatementEnd
