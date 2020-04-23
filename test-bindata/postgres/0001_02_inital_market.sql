-- +migrate Up
CREATE TABLE IF NOT EXISTS markets2
(
    id2             SERIAL PRIMARY KEY,
    symbol2         VARCHAR(32) UNIQUE,
    base_currency2  VARCHAR(8) NOT NULL DEFAULT '',
    quote_currency2 VARCHAR(8) NOT NULL DEFAULT ''
);

-- +migrate Down
DROP TABLE markets2;
