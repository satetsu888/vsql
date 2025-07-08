-- Test 15: Nested subqueries
-- Expected: 1 row (Tokyo)
-- Test: SELECT city, COUNT(*) as high_spenders FROM users WHERE id IN (SELECT user_id FROM orders WHERE amount > (SELECT AVG(amount) FROM orders)) GROUP BY city HAVING COUNT(*) > 1

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
SELECT city, COUNT(*) as high_spenders
FROM users
WHERE id IN (
    SELECT user_id 
    FROM orders 
    WHERE amount > (SELECT AVG(amount) FROM orders)
)
GROUP BY city
HAVING COUNT(*) > 1;

-- Cleanup
DROP TABLE orders;
DROP TABLE users;