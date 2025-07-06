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

-- Test JOIN queries
SELECT u.name, o.product, o.amount 
FROM users u 
INNER JOIN orders o ON u.id = o.user_id;

SELECT u.name, o.product 
FROM users u 
LEFT JOIN orders o ON u.id = o.user_id;

-- Test aggregate functions
SELECT COUNT(*) FROM users;
SELECT SUM(amount) FROM orders;
SELECT AVG(age) FROM users;
SELECT MAX(amount), MIN(amount) FROM orders;

-- Test GROUP BY
SELECT city, COUNT(*) as user_count 
FROM users 
GROUP BY city;

SELECT u.name, SUM(o.amount) as total_spent 
FROM users u 
JOIN orders o ON u.id = o.user_id 
GROUP BY u.name;

-- Test HAVING
SELECT city, COUNT(*) as count 
FROM users 
GROUP BY city 
HAVING COUNT(*) > 1;

-- Test subqueries
SELECT name 
FROM users 
WHERE id IN (SELECT user_id FROM orders WHERE amount > 500);

SELECT name, age 
FROM users 
WHERE age > (SELECT AVG(age) FROM users);

-- Test ORDER BY and LIMIT
SELECT * FROM users ORDER BY age DESC LIMIT 2;

-- Test complex query combining multiple features
SELECT u.city, COUNT(DISTINCT u.id) as users, SUM(o.amount) as total_sales
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
WHERE u.age >= 30
GROUP BY u.city
HAVING SUM(o.amount) > 0 OR SUM(o.amount) IS NULL
ORDER BY total_sales DESC;