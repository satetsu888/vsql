-- Test 11: UNION ALL queries
-- Expected: 8 rows (6 orders + 2 returns)
-- Test: SELECT 'order' as type, user_id, amount FROM orders UNION ALL SELECT 'return' as type, user_id, -amount FROM returns ORDER BY type, user_id

-- Create test tables
CREATE TABLE orders (
    id INTEGER,
    user_id INTEGER,
    product TEXT,
    amount INTEGER
);

CREATE TABLE returns (
    id INTEGER,
    user_id INTEGER,
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

INSERT INTO returns (id, user_id, amount) VALUES
    (1, 1, 50),
    (2, 3, 100);

-- Test query
SELECT 'order' as type, user_id, amount FROM orders
UNION ALL
SELECT 'return' as type, user_id, -amount FROM returns
ORDER BY type, user_id;

-- Cleanup
DROP TABLE returns;
DROP TABLE orders;