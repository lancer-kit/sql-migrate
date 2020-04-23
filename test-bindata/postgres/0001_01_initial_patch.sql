-- +migrate Up
CREATE TABLE IF NOT EXISTS markets1
(
    idtest             SERIAL PRIMARY KEY,
    symboltest         VARCHAR(32) UNIQUE,
    base_currencytest  VARCHAR(8) NOT NULL DEFAULT '',
    quote_currencytest VARCHAR(8) NOT NULL DEFAULT ''
);

-- +migrate Down
DROP TABLE markets1;
