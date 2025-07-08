-- Test 3: Comparison with NULL should return no rows
-- Expected: 0 rows
-- Bug result: may return rows where value is not NULL
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
SELECT * FROM null_test WHERE value = NULL;

-- Cleanup
DROP TABLE null_test;