-- Test: Boolean column storage
-- Status: Boolean values are not properly stored
-- Expected: true and false values
-- Actual: Empty or NULL values

CREATE TABLE bool_test (id int, active boolean);
INSERT INTO bool_test VALUES (1, true), (2, false);
SELECT id, active FROM bool_test;
DROP TABLE bool_test;