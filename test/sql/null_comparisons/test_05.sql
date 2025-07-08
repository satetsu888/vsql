-- Test 5: IS NULL operator (proper way to check NULL)
-- Expected: 1 row (id=2)

-- Setup
CREATE TABLE test_null (id INTEGER, val INTEGER);
INSERT INTO test_null VALUES (1, 100), (2, NULL), (3, 200);

-- Test Query
SELECT id, val FROM test_null WHERE val IS NULL;

-- Cleanup
DROP TABLE test_null;