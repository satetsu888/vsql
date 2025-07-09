-- Test 9: Multiple JOINs with complex WHERE
-- Testing multiple INNER JOINs with complex WHERE conditions

-- Setup
CREATE TABLE orders (id int, user_id int, product_id int, quantity int, price decimal, created_at timestamp);
CREATE TABLE products (id int, name text, category text, price decimal, stock int);
CREATE TABLE categories (id int, name text, parent_id int);

INSERT INTO categories VALUES 
  (1, 'Electronics', NULL),
  (2, 'Computers', 1),
  (3, 'Phones', 1),
  (4, 'Accessories', 1);

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
  o.id,
  o.user_id,
  p.name as product_name,
  c.name as category_name,
  o.quantity * o.price as order_total
FROM orders o
INNER JOIN products p ON o.product_id = p.id
INNER JOIN categories c ON p.category = c.name
WHERE o.created_at >= '2024-01-02'
  AND p.price > 50
  AND c.parent_id = 1
ORDER BY order_total DESC;

-- Cleanup
DROP TABLE orders;
DROP TABLE products;
DROP TABLE categories;