-- Test 7: JOIN with GROUP BY and HAVING
-- Expected: 2 rows (Alice: 1550, Charlie: 2400)
-- Status: FAILING
-- FAILING: HAVING clause with aggregate functions not implemented
-- Test: SELECT u.name, SUM(o.amount) as total_spent FROM users u JOIN orders o ON u.id = o.user_id GROUP BY u.name HAVING SUM(o.amount) > 1000 ORDER BY total_spent DESC

-- Create test tables
CREATE TABLE users (
    id INTEGER,
    name TEXT,
    age INTEGER,
    city TEXT
);

CREATE TABLE orders (
    id INTEGER,
    user_id INTEGER,
    product TEXT,
    amount INTEGER
);

-- Insert test data
INSERT INTO users (id, name, age, city) VALUES
    (1, 'Alice', 30, 'Tokyo'),
    (2, 'Bob', 25, 'Osaka'),
    (3, 'Charlie', 35, 'Tokyo'),
    (4, 'Dave', 45, 'Kyoto'),
    (5, 'Eve', 28, 'Osaka');

INSERT INTO orders (id, user_id, product, amount) VALUES
    (1, 1, 'Laptop', 1500),
    (2, 1, 'Mouse', 50),
    (3, 2, 'Keyboard', 100),
    (4, 3, 'Monitor', 800),
    (5, 3, 'Laptop', 1600),
    (6, 4, 'Mouse', 45);

-- Test query
SELECT u.name, SUM(o.amount) as total_spent
FROM users u
JOIN orders o ON u.id = o.user_id
GROUP BY u.name
HAVING SUM(o.amount) > 1000
ORDER BY total_spent DESC;

-- Cleanup
DROP TABLE orders;
DROP TABLE users;