-- Test: Boolean columns with AND operator in WHERE
-- Expected: 1 row

CREATE TABLE bool_expr (id int, flag1 boolean, flag2 boolean);
INSERT INTO bool_expr VALUES (1, true, true), (2, true, false), (3, false, false);

-- PostgreSQL: boolean columns can be used directly with AND in WHERE
-- This returns only id=1 (where both are true)
SELECT id FROM bool_expr WHERE flag1 AND flag2;

DROP TABLE bool_expr;