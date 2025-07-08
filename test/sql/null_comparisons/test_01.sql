-- Test 1: Direct NULL comparison with equals
-- Expected: 0 rows (NULL = NULL returns UNKNOWN, not TRUE)

-- Setup
CREATE TABLE test_null (id INTEGER, val INTEGER);
INSERT INTO test_null VALUES (1, 100), (2, NULL), (3, 200);

-- Test Query
SELECT id, val FROM test_null WHERE val = NULL;

-- Cleanup
DROP TABLE test_null;