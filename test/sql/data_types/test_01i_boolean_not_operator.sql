-- Test: NOT operator on boolean columns
-- Expected: 2 rows (false values)

CREATE TABLE bool_not (id int, active boolean);
INSERT INTO bool_not VALUES (1, true), (2, false), (3, true), (4, false);

SELECT id FROM bool_not WHERE NOT active;

DROP TABLE bool_not;