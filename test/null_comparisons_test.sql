-- Simple NULL Comparison Behavior Test
-- Tests basic NULL comparison behaviors in SQL

CREATE TABLE test_null (id INTEGER, val INTEGER);
INSERT INTO test_null VALUES (1, 100), (2, NULL), (3, 200);

-- Test 1: Direct NULL comparison with equals
-- Expected: 0 rows (NULL = NULL returns UNKNOWN, not TRUE)
SELECT id, val FROM test_null WHERE val = NULL;

-- Test 2: Direct NULL comparison with not equals
-- Expected: 0 rows (NULL != NULL returns UNKNOWN, not TRUE)
SELECT id, val FROM test_null WHERE val != NULL;

-- Test 3: Self comparison with NULLs
-- Expected: 2 rows (id=1,3 where val is not NULL)
-- NULL = NULL returns UNKNOWN, which is treated as FALSE in WHERE
SELECT id, val FROM test_null WHERE val = val;

-- Test 4: Self inequality with NULLs
-- Expected: 0 rows (even non-NULL values: 100 != 100 is FALSE)
-- NULL != NULL returns UNKNOWN, which is treated as FALSE
SELECT id, val FROM test_null WHERE val != val;

-- Test 5: IS NULL operator (proper way to check NULL)
-- Expected: 1 row (id=2)
SELECT id, val FROM test_null WHERE val IS NULL;

-- Test 6: IS NOT NULL operator
-- Expected: 2 rows (id=1,3)
SELECT id, val FROM test_null WHERE val IS NOT NULL;

-- Test 7: NULL with greater than comparison
-- Expected: 1 row (id=3 with val=200)
-- NULL > 100 returns UNKNOWN, treated as FALSE
SELECT id, val FROM test_null WHERE val > 100;

-- Test 8: NULL with less than comparison
-- Expected: 1 row (id=1 with val=100)
-- NULL < 150 returns UNKNOWN, treated as FALSE
SELECT id, val FROM test_null WHERE val < 150;

-- Test 9: COALESCE with NULL
-- Expected: 3 rows with val showing 100, 0, 200
SELECT id, COALESCE(val, 0) as val FROM test_null ORDER BY id;

-- Test 10: NULL in CASE expression
-- Expected: 3 rows with status showing 'has value', 'no value', 'has value'
SELECT id, 
       CASE 
           WHEN val IS NULL THEN 'no value'
           ELSE 'has value'
       END as status
FROM test_null
ORDER BY id;

DROP TABLE test_null;