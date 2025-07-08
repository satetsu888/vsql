-- Test 7: NULL with greater than comparison
-- Expected: 1 row (id=3 with val=200)
-- NULL > 100 returns UNKNOWN, treated as FALSE

-- Setup
CREATE TABLE test_null (id INTEGER, val INTEGER);
INSERT INTO test_null VALUES (1, 100), (2, NULL), (3, 200);

-- Test Query
SELECT id, val FROM test_null WHERE val > 100;

-- Cleanup
DROP TABLE test_null;