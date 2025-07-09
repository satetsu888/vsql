-- Test 23: NULL with BETWEEN operator
-- Expected: NULL comparisons with BETWEEN should return no match
-- Status: BETWEEN implemented, should handle NULL correctly

-- Setup
CREATE TABLE test_null (id INTEGER, val INTEGER);
INSERT INTO test_null VALUES 
    (1, 50),
    (2, NULL),
    (3, 150),
    (4, 250);

-- Test Query
SELECT id, val 
FROM test_null 
WHERE val BETWEEN 100 AND 200
ORDER BY id;

-- Cleanup
DROP TABLE test_null;