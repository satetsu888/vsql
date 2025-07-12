-- Test: NULL values and type inference
-- Expected: 4 rows

CREATE TABLE nullable_test (id text, value text);

-- First INSERT with NULL - type remains unknown
INSERT INTO nullable_test VALUES (1, NULL);    -- value: Unknown
-- Type is determined by first non-NULL value
INSERT INTO nullable_test VALUES (2, 100);     -- value: Unknown→Integer
-- NULL is always allowed after type is determined
INSERT INTO nullable_test VALUES (3, NULL);    -- value: Integer (NULL allowed)
-- Integer value is OK
INSERT INTO nullable_test VALUES (4, 200);     -- value: Integer

SELECT id, value FROM nullable_test ORDER BY id;

DROP TABLE nullable_test;