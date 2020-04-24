-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE balance (
    id int,
    balance int
);


-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE balance;
