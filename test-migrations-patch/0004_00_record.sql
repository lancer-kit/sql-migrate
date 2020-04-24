-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied
INSERT INTO people (id) VALUES (1);


-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
DELETE FROM people WHERE id=1;
