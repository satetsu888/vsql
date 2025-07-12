-- Test 9: MAX/MIN aggregates on numeric columns
-- Expected: 1 row
-- Test: Verifies MAX/MIN aggregates work correctly with numeric columns

-- Setup
CREATE TABLE numeric_test (
    id INTEGER,
    price DECIMAL,
    name TEXT,
    quantity INTEGER
);

INSERT INTO numeric_test (id, price, name, quantity) VALUES
    (1, 9.99, 'Item A', 100),
    (2, 10.01, 'Item B', 20),
    (10, 2.50, 'Item C', 5),
    (100, 99.99, 'Item D', 1000),
    (20, 5.00, 'Item E', 200);

-- Test Query
SELECT MAX(id) as max_id, MIN(id) as min_id FROM numeric_test;

-- Cleanup
DROP TABLE numeric_test;