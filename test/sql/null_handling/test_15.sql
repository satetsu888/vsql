-- Test 15: Aggregate functions with all NULL values
-- Expected: SUM/AVG/MAX/MIN should return NULL
-- Bug result: may return 0 or error

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

-- Test query
SELECT SUM(value) as sum_val, AVG(value) as avg_val, MAX(value) as max_val, MIN(value) as min_val 
FROM null_test WHERE name IS NULL;

-- Cleanup
DROP TABLE null_test;