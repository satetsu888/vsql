-- Test: Boolean column directly in WHERE clause (without comparison)
-- Expected: 2 rows (only true values)
-- Note: In standard SQL, WHERE <boolean_column> is equivalent to WHERE <boolean_column> = true

CREATE TABLE bool_direct (id int, active boolean);
INSERT INTO bool_direct VALUES (1, true), (2, false), (3, true), (4, false);

-- Should return only rows where active is true
SELECT id FROM bool_direct WHERE active;

DROP TABLE bool_direct;