-- Test 4: IS NULL operator (may not be implemented)
-- Expected: 2 rows (id=2,5)
-- Bug result: may fail or return wrong results

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
SELECT id, name FROM null_test WHERE value IS NULL;

-- Cleanup
DROP TABLE null_test;