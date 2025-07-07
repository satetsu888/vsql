-- Basic and Advanced Features Test
-- Converted from test_advanced.sql with expected values

-- Create tables for testing advanced features
CREATE TABLE users (id int, name text, age int, city text);
CREATE TABLE orders (id int, user_id int, product text, amount int);
CREATE TABLE products (id int, name text, price int);

-- Insert test data
INSERT INTO users (id, name, age, city) VALUES 
  (1, 'Alice', 25, 'Tokyo'),
  (2, 'Bob', 30, 'Osaka'),
  (3, 'Charlie', 35, 'Tokyo'),
  (4, 'David', 28, 'Kyoto');

INSERT INTO orders (id, user_id, product, amount) VALUES
  (1, 1, 'Laptop', 1200),
  (2, 1, 'Mouse', 50),
  (3, 2, 'Keyboard', 100),
  (4, 3, 'Monitor', 300),
  (5, 2, 'Laptop', 1500);

INSERT INTO products (id, name, price) VALUES
  (1, 'Laptop', 1000),
  (2, 'Mouse', 40),
  (3, 'Keyboard', 80),
  (4, 'Monitor', 250);

-- Test 1: INNER JOIN
-- Expected: 5 rows
SELECT u.name, o.product, o.amount 
FROM users u 
INNER JOIN orders o ON u.id = o.user_id;

-- Test 2: LEFT JOIN
-- Expected: 6 rows (including David with NULL product)
SELECT u.name, o.product 
FROM users u 
LEFT JOIN orders o ON u.id = o.user_id
ORDER BY u.name;

-- Test 3: COUNT aggregate
-- Expected: 1 row, count=4
SELECT COUNT(*) FROM users;

-- Test 4: SUM aggregate
-- Expected: 1 row, sum=3150
SELECT SUM(amount) FROM orders;

-- Test 5: AVG aggregate
-- Expected: 1 row, avg=29.5
SELECT AVG(age) FROM users;

-- Test 6: MAX and MIN aggregates
-- Expected: 1 row, max=1500, min=50
SELECT MAX(amount), MIN(amount) FROM orders;

-- Test 7: GROUP BY city
-- Expected: 3 rows (Tokyo: 2, Osaka: 1, Kyoto: 1)
SELECT city, COUNT(*) as user_count 
FROM users 
GROUP BY city
ORDER BY city;

-- Test 8: JOIN with GROUP BY and SUM
-- Expected: 3 rows (Alice: 1250, Bob: 1600, Charlie: 300)
SELECT u.name, SUM(o.amount) as total_spent 
FROM users u 
JOIN orders o ON u.id = o.user_id 
GROUP BY u.name
ORDER BY u.name;

-- Test 9: GROUP BY with HAVING
-- Expected: 1 row (Tokyo: 2)
SELECT city, COUNT(*) as count 
FROM users 
GROUP BY city 
HAVING COUNT(*) > 1;

-- Test 10: IN subquery
-- Expected: 2 rows (Alice, Bob)
SELECT name 
FROM users 
WHERE id IN (SELECT user_id FROM orders WHERE amount > 500)
ORDER BY name;

-- Test 11: Scalar subquery comparison
-- Expected: 2 rows (Bob: 30, Charlie: 35)
SELECT name, age 
FROM users 
WHERE age > (SELECT AVG(age) FROM users)
ORDER BY name;

-- Test 12: ORDER BY with LIMIT
-- Expected: 2 rows (Charlie: 35, Bob: 30)
SELECT * FROM users ORDER BY age DESC LIMIT 2;

-- Test 13: Complex query with multiple features
-- Expected: 3 rows
-- Tokyo: users=1, total_sales=300
-- Osaka: users=1, total_sales=1600
-- Kyoto: users=1, total_sales=NULL
SELECT u.city, COUNT(DISTINCT u.id) as users, SUM(o.amount) as total_sales
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
WHERE u.age >= 30
GROUP BY u.city
HAVING SUM(o.amount) > 0 OR SUM(o.amount) IS NULL
ORDER BY total_sales DESC;

-- Test 14: EXISTS subquery
-- Expected: 3 rows (Alice, Bob, Charlie)
SELECT name FROM users u
WHERE EXISTS (SELECT 1 FROM orders o WHERE o.user_id = u.id)
ORDER BY name;

-- Test 15: NOT EXISTS subquery
-- Expected: 1 row (David)
SELECT name FROM users u
WHERE NOT EXISTS (SELECT 1 FROM orders o WHERE o.user_id = u.id);

-- Cleanup
DROP TABLE products;
DROP TABLE orders;
DROP TABLE users;