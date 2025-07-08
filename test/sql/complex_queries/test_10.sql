-- Test 10: UNION with subqueries
-- Testing UNION operation with GROUP BY
-- Status: May fail - UNION not fully implemented

-- Setup
CREATE TABLE orders (id int, user_id int, product_id int, quantity int, price decimal, created_at timestamp);
CREATE TABLE products (id int, name text, category text, price decimal, stock int);

INSERT INTO products VALUES
  (1, 'Laptop', 'Computers', 1200.00, 50),
  (2, 'Mouse', 'Accessories', 25.00, 200),
  (3, 'Keyboard', 'Accessories', 75.00, 150),
  (4, 'iPhone', 'Phones', 999.00, 100),
  (5, 'Android Phone', 'Phones', 799.00, 80);

INSERT INTO orders VALUES
  (1, 101, 1, 1, 1200.00, '2024-01-01'),
  (2, 101, 2, 2, 50.00, '2024-01-02'),
  (3, 102, 1, 1, 1200.00, '2024-01-02'),
  (4, 102, 3, 1, 75.00, '2024-01-02'),
  (5, 103, 4, 1, 999.00, '2024-01-03'),
  (6, 103, 2, 3, 75.00, '2024-01-03');

-- Test Query
SELECT 'High Value' as order_type, user_id, SUM(price * quantity) as total
FROM orders
WHERE price * quantity > 1000
GROUP BY user_id
UNION
SELECT 'Low Value' as order_type, user_id, SUM(price * quantity) as total
FROM orders
WHERE price * quantity <= 1000
GROUP BY user_id
ORDER BY total DESC;

-- Cleanup
DROP TABLE orders;
DROP TABLE products;