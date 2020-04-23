-- +migrate Up
CREATE TABLE IF NOT EXISTS markets4
(
    id4             SERIAL PRIMARY KEY,
    symbol4         VARCHAR(32) UNIQUE,
    base_currency4  VARCHAR(8) NOT NULL DEFAULT '',
    quote_currency4 VARCHAR(8) NOT NULL DEFAULT ''
);

-- +migrate Down
DROP TABLE markets4;
