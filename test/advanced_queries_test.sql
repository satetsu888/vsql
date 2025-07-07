-- Advanced Query Test Cases for VSQL
-- Moved from examples/advanced_queries.sql with expected values added

-- Setup: Create test tables
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

CREATE TABLE products (
    name TEXT,
    price INTEGER,
    category TEXT
);

CREATE TABLE returns (
    id INTEGER,
    user_id INTEGER,
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

INSERT INTO products (name, price, category) VALUES
    ('Laptop', 1500, 'Electronics'),
    ('Mouse', 50, 'Accessories'),
    ('Keyboard', 100, 'Accessories'),
    ('Monitor', 800, 'Electronics');

INSERT INTO returns (id, user_id, amount) VALUES
    (1, 1, 50),
    (2, 3, 100);

-- Test 1: INNER JOIN
-- Expected: 6 rows
SELECT u.name, o.product, o.amount 
FROM users u 
INNER JOIN orders o ON u.id = o.user_id;

-- Test 2: LEFT JOIN with aggregation
-- Expected: 5 rows (all users, including Eve with 0 total)
-- FAILING: COALESCE function not implemented
SELECT u.name, COALESCE(SUM(o.amount), 0) as total_spent
FROM users u 
LEFT JOIN orders o ON u.id = o.user_id
GROUP BY u.name
ORDER BY u.name;

-- Test 3: IN subquery
-- Expected: 2 rows (Alice, Charlie)
SELECT name FROM users 
WHERE id IN (SELECT user_id FROM orders WHERE amount > 1000)
ORDER BY name;

-- Test 4: EXISTS subquery
-- Expected: 3 rows (Alice, Charlie, Dave)
-- FAILING: Test expects Dave but he only has amount=45 (< 500)
SELECT name FROM users u
WHERE EXISTS (SELECT 1 FROM orders o WHERE o.user_id = u.id AND o.amount > 500)
ORDER BY name;

-- Test 5: Scalar subquery in SELECT
-- Expected: 5 rows with correct order counts
SELECT name, 
       (SELECT COUNT(*) FROM orders o WHERE o.user_id = u.id) as order_count
FROM users u
ORDER BY name;

-- Test 6: GROUP BY with COUNT and AVG
-- Expected: 3 rows (Tokyo: 2 users, avg_age=32.5; Osaka: 2 users, avg_age=26.5; Kyoto: 1 user, avg_age=45)
SELECT city, COUNT(*) as user_count, AVG(age) as avg_age
FROM users 
GROUP BY city
ORDER BY city;

-- Test 7: JOIN with GROUP BY and HAVING
-- Expected: 2 rows (Alice: 1550, Charlie: 2400)
-- FAILING: ORDER BY with alias not working correctly
SELECT u.name, SUM(o.amount) as total_spent
FROM users u
JOIN orders o ON u.id = o.user_id
GROUP BY u.name
HAVING SUM(o.amount) > 1000
ORDER BY total_spent DESC;

-- Test 8: CTE (WITH clause) - Complex query
-- Expected: 3 rows with city statistics
-- FAILING: CTEs (WITH clause) not implemented
WITH city_stats AS (
  SELECT u.city, 
         COUNT(DISTINCT u.id) as user_count,
         COALESCE(SUM(o.amount), 0) as total_revenue,
         COALESCE(AVG(o.amount), 0) as avg_order_value
  FROM users u
  LEFT JOIN orders o ON u.id = o.user_id
  GROUP BY u.city
)
SELECT * FROM city_stats
WHERE total_revenue > 1000 OR user_count > 2
ORDER BY total_revenue DESC;

-- Test 9: Window Functions (if implemented)
-- Expected: 6 rows with running totals
-- Expected error if window functions not supported
-- FAILING: Window functions (OVER clause) not implemented
SELECT user_id, 
       amount,
       SUM(amount) OVER (PARTITION BY user_id ORDER BY id) as running_total
FROM orders
ORDER BY user_id, id;

-- Test 10: Multiple JOINs (three tables)
-- Expected: 6 rows
-- FAILING: Complex multi-table JOINs with column aliasing issues
SELECT u.name, o.amount, p.name as product_name, p.price
FROM users u
JOIN orders o ON u.id = o.user_id
JOIN products p ON o.product = p.name
ORDER BY u.name, o.amount;

-- Test 11: UNION ALL queries
-- Expected: 8 rows (6 orders + 2 returns)
SELECT 'order' as type, user_id, amount FROM orders
UNION ALL
SELECT 'return' as type, user_id, -amount FROM returns
ORDER BY type, user_id;

-- Test 12: Complex WHERE with BETWEEN and IN
-- Expected: 3 rows (Bob, Charlie, Dave)
-- FAILING: Expected 3 rows but query logic returns 4 rows (Alice, Bob, Charlie, Dave)
SELECT * FROM users
WHERE (age BETWEEN 25 AND 35 AND city IN ('Tokyo', 'Osaka'))
   OR (age > 40 AND city = 'Kyoto')
ORDER BY age DESC, name ASC;

-- Test 13: RIGHT JOIN (all orders, even without matching users)
-- Expected: 6 rows
SELECT u.name, o.product, o.amount
FROM users u
RIGHT JOIN orders o ON u.id = o.user_id
ORDER BY o.id;

-- Test 14: FULL OUTER JOIN (all users and all orders)
-- Expected: 7 rows (including unmatched)
SELECT COALESCE(u.name, 'Unknown') as name, 
       COALESCE(o.product, 'No Order') as product,
       COALESCE(o.amount, 0) as amount
FROM users u
FULL OUTER JOIN orders o ON u.id = o.user_id
ORDER BY u.id, o.id;

-- Test 15: Nested subqueries
-- Expected: 1 row (Tokyo)
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
DROP TABLE returns;
DROP TABLE products;
DROP TABLE orders;
DROP TABLE users;