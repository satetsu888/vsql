-- Test 24: NULL with LIKE operator
-- Expected: 2 rows
-- Status: LIKE implemented, should handle NULL correctly

-- Setup
CREATE TABLE test_null (id INTEGER, name TEXT);
INSERT INTO test_null VALUES 
    (1, 'Alice'),
    (2, NULL),
    (3, 'Bob'),
    (4, 'Bobby');

-- Test Query
SELECT id, name 
FROM test_null 
WHERE name LIKE 'B%'
ORDER BY id;

-- Cleanup
DROP TABLE test_null;