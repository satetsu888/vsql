-- Test 6: IS NOT NULL operator
-- Expected: 2 rows (id=1,3)

-- Setup
CREATE TABLE test_null (id INTEGER, val INTEGER);
INSERT INTO test_null VALUES (1, 100), (2, NULL), (3, 200);

-- Test Query
SELECT id, val FROM test_null WHERE val IS NOT NULL;

-- Cleanup
DROP TABLE test_null;