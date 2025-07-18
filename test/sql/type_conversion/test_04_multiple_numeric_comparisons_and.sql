-- Test 4: Multiple numeric comparisons
-- Expected: 1 row (id=20)
-- Test: Verifies multiple numeric comparisons with AND work correctly

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
SELECT * FROM numeric_test WHERE id > 10 AND quantity < 300;

-- Cleanup
DROP TABLE numeric_test;