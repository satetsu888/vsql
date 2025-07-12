-- Test: COUNT DISTINCT function - With JOIN
-- Expected: 1 row
-- Test: Verifies COUNT DISTINCT works correctly with JOINs

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

-- Test Query - COUNT DISTINCT with JOIN
SELECT COUNT(DISTINCT u.id) as unique_users_with_orders
FROM users u
INNER JOIN orders o ON u.id = o.user_id;

-- Cleanup
DROP TABLE orders;
DROP TABLE users;