-- Test: Boolean literals in SELECT without table
-- Expected: 1 row
-- PostgreSQL allows SELECT without FROM clause

SELECT true AS bool_val, 'true literal' AS description;