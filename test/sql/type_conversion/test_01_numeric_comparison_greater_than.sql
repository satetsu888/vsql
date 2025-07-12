-- Test 1: Numeric comparison - should return 3 rows where id > 5
-- Expected: 3 rows with id=10,20,100 (values > 5)
-- Test: Verifies numeric comparison works correctly, not string comparison

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
SELECT id, name FROM numeric_test WHERE id > 5 ORDER BY id;

-- Cleanup
DROP TABLE numeric_test;