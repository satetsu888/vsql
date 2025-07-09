-- Test 12: GROUP BY with NULLs
-- Expected: NULL values should be grouped together
-- Bug result: may handle NULL grouping incorrectly

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
SELECT name, COUNT(*) as count FROM null_test GROUP BY name ORDER BY name;

-- Cleanup
DROP TABLE null_test;