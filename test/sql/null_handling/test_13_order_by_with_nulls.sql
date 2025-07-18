-- Test 13: ORDER BY with NULLs
-- Expected: NULLs should be sorted consistently (either first or last)
-- Bug result: may sort inconsistently
-- Expected: 5 rows

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
SELECT id, name FROM null_test ORDER BY name;

-- Cleanup
DROP TABLE null_test;