-- Test: Boolean literals directly in WHERE clause
-- Expected: 1 row

CREATE TABLE bool_literal (id int, name text);
INSERT INTO bool_literal VALUES (1, 'test');

-- This should return the row because true is always true
SELECT id FROM bool_literal WHERE true;

DROP TABLE bool_literal;