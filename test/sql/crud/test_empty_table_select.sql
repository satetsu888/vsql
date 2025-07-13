-- Test: SELECT from empty table should show column headers
-- Expected: 0 rows
-- PostgreSQL shows column headers even for empty result sets

CREATE TABLE empty_users (id int, name text, email text);

-- Should show column headers with (0 rows)
SELECT id, name FROM empty_users;

DROP TABLE empty_users;