-- +migrate Up
CREATE TABLE IF NOT EXISTS markets3
(
    id3             SERIAL PRIMARY KEY,
    symbol3         VARCHAR(32) UNIQUE,
    base_currency3  VARCHAR(8) NOT NULL DEFAULT '',
    quote_currency3 VARCHAR(8) NOT NULL DEFAULT ''
);

-- +migrate Down
DROP TABLE markets3;
