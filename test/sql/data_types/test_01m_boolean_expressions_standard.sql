-- Test: Standard SQL boolean expressions
-- Expected: error
-- Status: FAILING
-- Note: These standard SQL boolean expressions are not properly handled

CREATE TABLE bool_expr (id int, flag1 boolean, flag2 boolean);
INSERT INTO bool_expr VALUES (1, true, true), (2, true, false), (3, false, false);

-- Standard SQL: boolean columns can be used directly in WHERE
-- This should return only id=1 (where both are true)
SELECT id FROM bool_expr WHERE flag1 AND flag2;

DROP TABLE bool_expr;