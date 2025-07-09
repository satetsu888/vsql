-- Test 4: Subquery in FROM clause (derived table)
-- Testing derived tables in FROM clause
-- Status: May fail - derived tables not fully supported

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
  category,
  AVG(total_sales) as avg_category_sales
FROM (
  SELECT 
    p.category,
    p.id,
    SUM(o.price * o.quantity) as total_sales
  FROM products p
  INNER JOIN orders o ON p.id = o.product_id
  GROUP BY p.category, p.id
) as product_sales
GROUP BY category;

-- Cleanup
DROP TABLE orders;
DROP TABLE products;