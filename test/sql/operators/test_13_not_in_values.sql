-- Test 13: NOT IN with value list (no NULLs)
-- Expected: 2 rows (id=2,20)
-- Test: IDs not in the list (1, 10, 100)

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
SELECT id, name FROM numeric_test WHERE id NOT IN (1, 10, 100);

-- Cleanup
DROP TABLE numeric_test;