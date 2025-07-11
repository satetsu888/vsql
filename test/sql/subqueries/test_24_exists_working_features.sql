-- EXISTS/NOT EXISTS Working Features Test
-- This file tests only the features that are currently working

-- Setup
CREATE TABLE customers (id int, name text, country text);
CREATE TABLE orders (id int, customer_id int, total int);

INSERT INTO customers VALUES
  (1, 'Alice', 'Japan'),
  (2, 'Bob', 'USA'),
  (3, 'Charlie', 'UK'),
  (4, 'David', 'Japan');

INSERT INTO orders VALUES
  (1, 1, 100),
  (2, 1, 200),
  (3, 2, 150),
  (4, 4, 300);

-- Test 1: Basic correlated EXISTS
-- Expected: 3 rows (Alice, Bob, David - customers with orders)
SELECT name FROM customers c
WHERE EXISTS (SELECT 1 FROM orders o WHERE o.customer_id = c.id)
ORDER BY name;

-- Test 2: Basic correlated NOT EXISTS  
-- Expected: 1 row (Charlie - customer without orders)
SELECT name FROM customers c
WHERE NOT EXISTS (SELECT 1 FROM orders o WHERE o.customer_id = c.id);

-- Test 3: Non-correlated EXISTS
-- Expected: 4 rows (all customers, because orders table has rows)
SELECT name FROM customers
WHERE EXISTS (SELECT 1 FROM orders WHERE total > 0)
ORDER BY name;

-- Test 4: EXISTS with multiple conditions
-- Expected: 2 rows (Alice, David - Japanese customers with orders)
SELECT name FROM customers c
WHERE country = 'Japan'
AND EXISTS (SELECT 1 FROM orders o WHERE o.customer_id = c.id)
ORDER BY name;

-- Test 5: EXISTS with WHERE clause in subquery
-- Expected: 2 rows (Alice, David - customers with orders > 150)
SELECT name FROM customers c
WHERE EXISTS (SELECT 1 FROM orders o WHERE o.customer_id = c.id AND o.total > 150)
ORDER BY name;

-- Test 6: NOT EXISTS with condition
-- Expected: 3 rows (Bob, Charlie, David - customers without small orders)
SELECT name FROM customers c
WHERE NOT EXISTS (SELECT 1 FROM orders o WHERE o.customer_id = c.id AND o.total < 150)
ORDER BY name;

-- Test 7: Combined EXISTS and NOT EXISTS
-- Expected: 1 row (David - has large order but no small order)
SELECT name FROM customers c
WHERE EXISTS (SELECT 1 FROM orders o WHERE o.customer_id = c.id AND o.total > 200)
AND NOT EXISTS (SELECT 1 FROM orders o WHERE o.customer_id = c.id AND o.total < 200)
ORDER BY name;

-- Cleanup
DROP TABLE orders;
DROP TABLE customers;