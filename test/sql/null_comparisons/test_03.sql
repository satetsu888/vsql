-- Test 3: Self comparison with NULLs
-- Expected: 2 rows (id=1,3 where val is not NULL)
-- NULL = NULL returns UNKNOWN, which is treated as FALSE in WHERE

-- Setup
CREATE TABLE test_null (id INTEGER, val INTEGER);
INSERT INTO test_null VALUES (1, 100), (2, NULL), (3, 200);

-- Test Query
SELECT id, val FROM test_null WHERE val = val;

-- Cleanup
DROP TABLE test_null;