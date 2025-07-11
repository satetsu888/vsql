-- Test 13: Complex query with multiple features
-- Expected: 3 rows
-- Tokyo: users=1, total_sales=300
-- Osaka: users=1, total_sales=1600
-- Kyoto: users=1, total_sales=NULL
-- Test: Verifies GROUP BY with LEFT JOIN, HAVING clause with OR condition and IS NULL check

-- Setup
CREATE TABLE users (id int, name text, age int, city text);
CREATE TABLE orders (id int, user_id int, product text, amount int);

INSERT INTO users (id, name, age, city) VALUES 
  (1, 'Alice', 25, 'Tokyo'),
  (2, 'Bob', 30, 'Osaka'),
  (3, 'Charlie', 35, 'Tokyo'),
  (4, 'David', 30, 'Kyoto');

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
HAVING SUM(o.amount) > 0 OR SUM(o.amount) IS NULL
ORDER BY total_sales DESC;

-- Cleanup
DROP TABLE orders;
DROP TABLE users;