-- Test 10: NULL in NOT IN clause
-- Expected: 0 rows (NOT IN with NULL returns UNKNOWN for all non-NULL values)
-- Test: Verifies NOT IN with NULL in list returns no rows

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
SELECT id, value FROM null_test WHERE value NOT IN (0, 100, NULL);

-- Cleanup
DROP TABLE null_test;