-- Test 11: COUNT with NULLs
-- Expected: COUNT(*) = 5, COUNT(value) = 3
-- Bug result: may count NULLs incorrectly
-- Expected: 1 row

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
SELECT COUNT(*) as total_rows, COUNT(value) as non_null_values FROM null_test;

-- Cleanup
DROP TABLE null_test;