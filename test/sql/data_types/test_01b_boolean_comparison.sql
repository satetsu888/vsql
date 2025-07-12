-- Test: Boolean comparison
-- Expected: 1 row (id=1 with true)
-- Status: FAILING - Boolean values are not properly compared, returns 0 rows

CREATE TABLE bool_compare (id int, active boolean);
INSERT INTO bool_compare VALUES (1, true), (2, false);
SELECT id FROM bool_compare WHERE active = true;
DROP TABLE bool_compare;