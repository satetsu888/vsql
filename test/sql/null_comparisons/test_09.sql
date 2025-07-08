-- Test 9: COALESCE with NULL
-- Expected: 3 rows with val showing 100, 0, 200
-- Status: May fail - COALESCE not implemented

-- Setup
CREATE TABLE test_null (id INTEGER, val INTEGER);
INSERT INTO test_null VALUES (1, 100), (2, NULL), (3, 200);

-- Test Query
SELECT id, COALESCE(val, 0) as val FROM test_null ORDER BY id;

-- Cleanup
DROP TABLE test_null;