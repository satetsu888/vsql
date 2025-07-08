-- Test 10: NULL in CASE expression
-- Expected: 3 rows with status showing 'has value', 'no value', 'has value'
-- Status: May fail - CASE expressions not implemented

-- Setup
CREATE TABLE test_null (id INTEGER, val INTEGER);
INSERT INTO test_null VALUES (1, 100), (2, NULL), (3, 200);

-- Test Query
SELECT id, 
       CASE 
           WHEN val IS NULL THEN 'no value'
           ELSE 'has value'
       END as status
FROM test_null
ORDER BY id;

-- Cleanup
DROP TABLE test_null;