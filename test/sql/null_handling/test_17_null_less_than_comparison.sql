-- Test 17: NULL with less than comparison
-- Expected: 1 row (id=1 with val=100)
-- NULL < 150 returns UNKNOWN, treated as FALSE

-- Setup
CREATE TABLE test_null (id INTEGER, val INTEGER);
INSERT INTO test_null VALUES (1, 100), (2, NULL), (3, 200);

-- Test Query
SELECT id, val FROM test_null WHERE val < 150;

-- Cleanup
DROP TABLE test_null;