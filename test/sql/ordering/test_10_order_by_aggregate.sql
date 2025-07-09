-- Test 10: ORDER BY aggregate value
-- Expected: 3 rows ordered by total_sales DESC: Osaka (1600), Tokyo (300), Kyoto (NULL)
-- Based on: basic_advanced/test_13.sql

-- Setup
CREATE TABLE users (id int, name text, age int, city text);
CREATE TABLE orders (id int, user_id int, product text, amount int);

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

-- Test Query
SELECT u.city, COUNT(DISTINCT u.id) as users, SUM(o.amount) as total_sales
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
WHERE u.age >= 30
GROUP BY u.city
ORDER BY total_sales DESC;

-- Cleanup
DROP TABLE orders;
DROP TABLE users;