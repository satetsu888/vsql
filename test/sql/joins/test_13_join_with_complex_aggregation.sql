-- Test 11: Complex aggregation with HAVING
-- Testing multiple aggregate functions with GROUP BY and HAVING

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
SELECT 
  p.category,
  COUNT(DISTINCT o.user_id) as unique_customers,
  COUNT(o.id) as total_orders,
  SUM(o.quantity) as total_quantity,
  SUM(o.price * o.quantity) as total_revenue,
  AVG(o.price * o.quantity) as avg_order_value
FROM products p
INNER JOIN orders o ON p.id = o.product_id
GROUP BY p.category
HAVING SUM(o.quantity) > 1 
  AND COUNT(DISTINCT o.user_id) >= 2;
-- Expected: 2 rows

-- Cleanup
DROP TABLE orders;
DROP TABLE products;