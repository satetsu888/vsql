-- Test 9: Aggregate functions with all NULL values
-- Expected: SUM/AVG/MAX/MIN should return NULL, COUNT should return 0

-- Create test table
CREATE TABLE null_test (
    id INTEGER,
    name TEXT,
    value INTEGER,
    description TEXT
);

-- Insert test data with all NULL values in the value column
INSERT INTO null_test (id, name, value, description) VALUES
    (1, 'Alice', NULL, 'No value'),
    (2, 'Bob', NULL, 'No value'),
    (3, 'Charlie', NULL, 'No value');

-- Test query: All aggregates on NULL-only column
-- Expected: 1 row
SELECT 
    COUNT(value) as count_val,  -- Should be 0
    SUM(value) as sum_val,      -- Should be NULL
    AVG(value) as avg_val,      -- Should be NULL
    MAX(value) as max_val,      -- Should be NULL
    MIN(value) as min_val       -- Should be NULL
FROM null_test;

-- Cleanup
DROP TABLE null_test;