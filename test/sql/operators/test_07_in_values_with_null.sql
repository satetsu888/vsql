-- Test 9: NULL in IN clause
-- Expected: 2 rows (id=1,4) - NULL in list doesn't match
-- Bug result: may return wrong results
-- Status: FAILING

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
SELECT id, value FROM null_test WHERE value IN (0, 100, NULL);

-- Cleanup
DROP TABLE null_test;