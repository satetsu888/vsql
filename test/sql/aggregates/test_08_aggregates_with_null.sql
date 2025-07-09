-- Test 8: Aggregate functions with NULL values
-- Expected: COUNT(*) = 5, COUNT(value) = 3

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

-- Test query 1: COUNT with NULLs
SELECT COUNT(*) as total_rows, COUNT(value) as non_null_values FROM null_test;

-- Test query 2: Other aggregates with NULLs (should ignore NULL values)
SELECT 
    SUM(value) as sum_val,     -- Should be 300 (100+200+0)
    AVG(value) as avg_val,     -- Should be 100 (300/3)
    MAX(value) as max_val,     -- Should be 200
    MIN(value) as min_val      -- Should be 0
FROM null_test;

-- Cleanup
DROP TABLE null_test;