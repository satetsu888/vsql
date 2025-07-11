-- Test: NOT EXISTS with condition
-- Expected: 3 rows (Bob, Charlie, David - customers without small orders)

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
WHERE NOT EXISTS (SELECT 1 FROM orders o WHERE o.customer_id = c.id AND o.total < 150)
ORDER BY name;

-- Cleanup
DROP TABLE orders;
DROP TABLE customers;