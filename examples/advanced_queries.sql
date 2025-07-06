-- Advanced Query Examples for VSQL

-- 1. JOIN Examples
-- INNER JOIN
SELECT u.name, o.product, o.amount 
FROM users u 
INNER JOIN orders o ON u.id = o.user_id;

-- LEFT JOIN (all users, even without orders)
SELECT u.name, COALESCE(SUM(o.amount), 0) as total_spent
FROM users u 
LEFT JOIN orders o ON u.id = o.user_id
GROUP BY u.name;

-- 2. Subquery Examples
-- IN subquery
SELECT name FROM users 
WHERE id IN (SELECT user_id FROM orders WHERE amount > 1000);

-- EXISTS subquery
SELECT name FROM users u
WHERE EXISTS (SELECT 1 FROM orders o WHERE o.user_id = u.id AND o.amount > 500);

-- Scalar subquery in SELECT
SELECT name, 
       (SELECT COUNT(*) FROM orders o WHERE o.user_id = u.id) as order_count
FROM users u;

-- 3. Aggregate Functions with GROUP BY
-- Count users by city
SELECT city, COUNT(*) as user_count, AVG(age) as avg_age
FROM users 
GROUP BY city;

-- Top spending users
SELECT u.name, SUM(o.amount) as total_spent
FROM users u
JOIN orders o ON u.id = o.user_id
GROUP BY u.name
HAVING SUM(o.amount) > 1000
ORDER BY total_spent DESC;

-- 4. Complex Combined Query
-- Find cities with high-value customers
WITH city_stats AS (
  SELECT u.city, 
         COUNT(DISTINCT u.id) as user_count,
         SUM(o.amount) as total_revenue,
         AVG(o.amount) as avg_order_value
  FROM users u
  LEFT JOIN orders o ON u.id = o.user_id
  GROUP BY u.city
)
SELECT * FROM city_stats
WHERE total_revenue > 1000 OR user_count > 2
ORDER BY total_revenue DESC;

-- 5. Window Functions (if implemented)
-- Running total of orders
SELECT user_id, 
       amount,
       SUM(amount) OVER (PARTITION BY user_id ORDER BY id) as running_total
FROM orders;

-- 6. Multiple JOINs
-- Users, their orders, and product details
SELECT u.name, o.amount, p.name as product_name, p.price
FROM users u
JOIN orders o ON u.id = o.user_id
JOIN products p ON o.product = p.name;

-- 7. UNION queries
-- All transactions (both orders and returns)
SELECT 'order' as type, user_id, amount FROM orders
UNION ALL
SELECT 'return' as type, user_id, -amount FROM returns;

-- 8. Complex WHERE conditions
-- Users with specific criteria
SELECT * FROM users
WHERE (age BETWEEN 25 AND 35 AND city IN ('Tokyo', 'Osaka'))
   OR (age > 40 AND city = 'Kyoto')
ORDER BY age DESC, name ASC;