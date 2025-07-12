-- Test: Boolean column storage
-- Expected: 2 rows (with true and false values)

CREATE TABLE bool_test (id int, active boolean);
INSERT INTO bool_test VALUES (1, true), (2, false);
SELECT id, active FROM bool_test;
DROP TABLE bool_test;