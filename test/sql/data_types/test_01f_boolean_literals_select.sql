-- Test: Boolean literals in SELECT without table
-- Expected: error
-- Status: FAILING
-- Note: SELECT without FROM clause is not supported

SELECT true AS bool_val, 'true literal' AS description;