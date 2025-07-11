-- Test: Combined EXISTS and NOT EXISTS
-- Expected: 1 row (David - has large order but no small order)

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

-- Test query
SELECT name FROM customers c
WHERE EXISTS (SELECT 1 FROM orders o WHERE o.customer_id = c.id AND o.total > 200)
AND NOT EXISTS (SELECT 1 FROM orders o WHERE o.customer_id = c.id AND o.total < 200)
ORDER BY name;

-- Cleanup
DROP TABLE orders;
DROP TABLE customers;