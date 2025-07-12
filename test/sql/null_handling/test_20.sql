-- Test 20: NULL in arithmetic operations
-- Expected: All results should be NULL (NULL propagates through arithmetic)
-- Status: May fail if NULL arithmetic not properly implemented
-- Expected: 3 rows

-- Setup
CREATE TABLE test_null (id INTEGER, val INTEGER);
INSERT INTO test_null VALUES (1, 100), (2, NULL), (3, 200);

-- Test Query
SELECT 
    id,
    val,
    val + 10 as add_result,
    val - 10 as sub_result,
    val * 2 as mul_result,
    val / 2 as div_result
FROM test_null
ORDER BY id;

-- Cleanup
DROP TABLE test_null;