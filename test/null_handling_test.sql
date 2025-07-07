-- Test for NULL handling bugs
-- Bug: NULL comparisons don't follow SQL three-valued logic

-- Create test table
CREATE TABLE null_test (
    id INTEGER,
    name TEXT,
    value INTEGER,
    description TEXT
);

-- Insert test data including NULLs
INSERT INTO null_test (id, name, value, description) VALUES
    (1, 'Alice', 100, 'Has value'),
    (2, 'Bob', NULL, 'No value'),
    (3, NULL, 200, 'No name'),
    (4, 'Charlie', 0, 'Zero value'),
    (5, NULL, NULL, NULL);

-- Test 1: NULL = NULL should return no rows (in SQL, NULL = NULL is UNKNOWN, not TRUE)
-- Expected: 3 rows
-- Bug result: may return rows where both values are NULL
SELECT * FROM null_test WHERE name = name;

-- Test 2: NULL != NULL should also return no rows
-- Expected: 0 rows  
-- Bug result: may return rows
SELECT * FROM null_test WHERE value != value;

-- Test 3: Comparison with NULL should return no rows
-- Expected: 0 rows
-- Bug result: may return rows where value is not NULL
SELECT * FROM null_test WHERE value = NULL;

-- Test 4: IS NULL operator (may not be implemented)
-- Expected: 2 rows (id=2,5)
-- Bug result: may fail or return wrong results
SELECT id, name FROM null_test WHERE value IS NULL;

-- Test 5: IS NOT NULL operator
-- Expected: 3 rows (id=1,3,4)
-- Bug result: may fail or return wrong results
SELECT id, name FROM null_test WHERE value IS NOT NULL;

-- Test 6: NULL in arithmetic comparisons
-- Expected: 2 rows (id=1,3) - NULL comparisons should not match
-- Bug result: may include NULL values
SELECT id, value FROM null_test WHERE value > 50;

-- Test 7: NULL with AND operator
-- Expected: 0 rows (NULL AND TRUE = UNKNOWN)
-- Bug result: may return rows
SELECT * FROM null_test WHERE value > 50 AND name = NULL;

-- Test 8: NULL with OR operator  
-- Expected: 2 rows (id=1,3) - TRUE OR UNKNOWN = TRUE
-- Bug result: may return wrong results
SELECT id FROM null_test WHERE value > 50 OR name = NULL;

-- Test 9: NULL in IN clause
-- Expected: 2 rows (id=1,4) - NULL in list doesn't match
-- Bug result: may return wrong results
SELECT id, value FROM null_test WHERE value IN (0, 100, NULL);

-- Test 10: NULL in NOT IN clause
-- Expected: 0 rows (NOT IN with NULL returns UNKNOWN for all non-NULL values)
-- Bug result: may return rows
SELECT id, value FROM null_test WHERE value NOT IN (0, 100, NULL);

-- Test 11: COUNT with NULLs
-- Expected: COUNT(*) = 5, COUNT(value) = 3
-- Bug result: may count NULLs incorrectly
SELECT COUNT(*) as total_rows, COUNT(value) as non_null_values FROM null_test;

-- Test 12: GROUP BY with NULLs
-- Expected: NULL values should be grouped together
-- Bug result: may handle NULL grouping incorrectly
SELECT name, COUNT(*) as count FROM null_test GROUP BY name ORDER BY name;

-- Test 13: ORDER BY with NULLs
-- Expected: NULLs should be sorted consistently (either first or last)
-- Bug result: may sort inconsistently
SELECT id, name FROM null_test ORDER BY name;

-- Test 14: DISTINCT with NULLs
-- Expected: Should treat all NULLs as one distinct value
-- Bug result: may handle incorrectly
SELECT DISTINCT name FROM null_test ORDER BY name;

-- Test 15: Aggregate functions with all NULL values
-- Expected: SUM/AVG/MAX/MIN should return NULL
-- Bug result: may return 0 or error
SELECT SUM(value) as sum_val, AVG(value) as avg_val, MAX(value) as max_val, MIN(value) as min_val 
FROM null_test WHERE name IS NULL;

-- Cleanup
DROP TABLE null_test;