-- +migrate Up
CREATE TABLE IF NOT EXISTS markets
(
    id             SERIAL PRIMARY KEY,
    symbol         VARCHAR(32) UNIQUE,
    base_currency  VARCHAR(8) NOT NULL DEFAULT '',
    quote_currency VARCHAR(8) NOT NULL DEFAULT ''
);

-- +migrate Down
DROP TABLE markets;
