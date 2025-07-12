-- Test: Boolean with IS NULL - Shows that boolean values might be stored as NULL
-- Expected: 3 rows

CREATE TABLE bool_null (id int, active boolean);
INSERT INTO bool_null VALUES (1, true), (2, false), (3, NULL);
SELECT id, active IS NULL as is_null FROM bool_null;
DROP TABLE bool_null;