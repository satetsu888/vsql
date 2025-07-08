-- Test 4: Self inequality with NULLs
-- Expected: 0 rows (even non-NULL values: 100 != 100 is FALSE)
-- NULL != NULL returns UNKNOWN, which is treated as FALSE

-- Setup
CREATE TABLE test_null (id INTEGER, val INTEGER);
INSERT INTO test_null VALUES (1, 100), (2, NULL), (3, 200);

-- Test Query
SELECT id, val FROM test_null WHERE val != val;

-- Cleanup
DROP TABLE test_null;