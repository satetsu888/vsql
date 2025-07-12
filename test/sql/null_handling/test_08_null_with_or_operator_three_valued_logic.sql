-- Test 8: NULL with OR operator  
-- Expected: 2 rows (id=1,3) - TRUE OR UNKNOWN = TRUE
-- Test: Verifies OR with NULL comparison handles three-valued logic correctly

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
SELECT id FROM null_test WHERE value > 50 OR name = NULL;

-- Cleanup
DROP TABLE null_test;