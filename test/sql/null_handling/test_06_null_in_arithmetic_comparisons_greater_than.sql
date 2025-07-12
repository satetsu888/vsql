-- Test 6: NULL in arithmetic comparisons
-- Expected: 2 rows (id=1,3) - NULL comparisons should not match
-- Test: Verifies NULL values are correctly excluded from comparison operations

-- Setup
CREATE TABLE null_test (
    id INTEGER,
    name TEXT,
    value INTEGER,
    description TEXT
);

INSERT INTO null_test (id, name, value, description) VALUES
    (1, 'Alice', 100, 'Has value'),
    (2, 'Bob', NULL, 'No value'),
    (3, NULL, 200, 'No name'),
    (4, 'Charlie', 0, 'Zero value'),
    (5, NULL, NULL, NULL);

-- Test Query
SELECT id, value FROM null_test WHERE value > 50;

-- Cleanup
DROP TABLE null_test;