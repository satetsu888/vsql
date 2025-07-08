-- Test 9: Window Functions (if implemented)
-- Expected: 6 rows with running totals
-- Expected error if window functions not supported
-- Status: FAILING
-- FAILING: Window functions (OVER clause) not implemented
-- Test: SELECT user_id, amount, SUM(amount) OVER (PARTITION BY user_id ORDER BY id) as running_total FROM orders ORDER BY user_id, id

-- Create test tables
CREATE TABLE orders (
    id INTEGER,
    user_id INTEGER,
    product TEXT,
    amount INTEGER
);

-- Insert test data
INSERT INTO orders (id, user_id, product, amount) VALUES
    (1, 1, 'Laptop', 1500),
    (2, 1, 'Mouse', 50),
    (3, 2, 'Keyboard', 100),
    (4, 3, 'Monitor', 800),
    (5, 3, 'Laptop', 1600),
    (6, 4, 'Mouse', 45);

-- Test query
SELECT user_id, 
       amount,
       SUM(amount) OVER (PARTITION BY user_id ORDER BY id) as running_total
FROM orders
ORDER BY user_id, id;

-- Cleanup
DROP TABLE orders;