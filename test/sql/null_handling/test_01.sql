-- Test 1: NULL = NULL should return no rows (in SQL, NULL = NULL is UNKNOWN, not TRUE)
-- Expected: 3 rows
-- Test: Verifies that NULL = NULL evaluates to UNKNOWN, not TRUE

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
SELECT * FROM null_test WHERE name = name;

-- Cleanup
DROP TABLE null_test;