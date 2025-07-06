-- Complex Query Tests for VSQL
-- Testing JOINs, subqueries, and their combinations

-- Test Setup
CREATE TABLE orders (id int, user_id int, product_id int, quantity int, price decimal, created_at timestamp);
CREATE TABLE products (id int, name text, category text, price decimal, stock int);
CREATE TABLE categories (id int, name text, parent_id int);

-- Insert test data
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

-- Test 1: Complex JOIN with multiple tables
SELECT 
  o.id as order_id,
  o.user_id,
  p.name as product_name,
  p.category,
  o.quantity,
  o.price * o.quantity as total_price
FROM orders o
INNER JOIN products p ON o.product_id = p.id
WHERE p.category IN ('Computers', 'Phones')
ORDER BY o.created_at DESC;

-- Test 2: LEFT JOIN with NULL handling
SELECT 
  p.name,
  p.stock,
  COALESCE(SUM(o.quantity), 0) as total_ordered
FROM products p
LEFT JOIN orders o ON p.id = o.product_id
GROUP BY p.id, p.name, p.stock
HAVING p.stock > 0;

-- Test 3: Subquery in SELECT clause
SELECT 
  p.name,
  p.price,
  (SELECT COUNT(*) FROM orders o WHERE o.product_id = p.id) as order_count,
  (SELECT SUM(o.quantity) FROM orders o WHERE o.product_id = p.id) as total_quantity
FROM products p
WHERE p.category = 'Accessories';

-- Test 4: Subquery in FROM clause (derived table)
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

-- Test 5: EXISTS subquery
SELECT * FROM products p
WHERE EXISTS (
  SELECT 1 FROM orders o 
  WHERE o.product_id = p.id 
  AND o.quantity > 1
);

-- Test 6: NOT EXISTS subquery
SELECT * FROM products p
WHERE NOT EXISTS (
  SELECT 1 FROM orders o 
  WHERE o.product_id = p.id
);

-- Test 7: IN subquery with aggregation
SELECT * FROM products
WHERE id IN (
  SELECT product_id 
  FROM orders 
  GROUP BY product_id 
  HAVING SUM(quantity) > 2
);

-- Test 8: Correlated subquery with comparison
SELECT 
  p1.name,
  p1.price
FROM products p1
WHERE p1.price > (
  SELECT AVG(p2.price)
  FROM products p2
  WHERE p2.category = p1.category
);

-- Test 9: Multiple JOINs with complex WHERE
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

-- Test 10: UNION with subqueries
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

-- Test 11: Complex aggregation with HAVING
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

-- Test 12: Window functions (if supported)
SELECT 
  user_id,
  product_id,
  quantity,
  price,
  SUM(price * quantity) OVER (PARTITION BY user_id) as user_total,
  ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY price DESC) as price_rank
FROM orders;

-- Test 13: Self-join
SELECT 
  c1.name as category,
  c2.name as parent_category
FROM categories c1
LEFT JOIN categories c2 ON c1.parent_id = c2.id;

-- Test 14: Complex CASE statements
SELECT 
  name,
  price,
  CASE 
    WHEN price < 100 THEN 'Budget'
    WHEN price BETWEEN 100 AND 500 THEN 'Mid-range'
    WHEN price > 500 THEN 'Premium'
  END as price_category,
  CASE category
    WHEN 'Phones' THEN 'Mobile Devices'
    WHEN 'Computers' THEN 'Computing'
    ELSE 'Other'
  END as category_group
FROM products;

-- Test 15: ALL/ANY subqueries
SELECT * FROM products
WHERE price > ALL (
  SELECT price FROM products WHERE category = 'Accessories'
);

SELECT * FROM products
WHERE price > ANY (
  SELECT price FROM products WHERE category = 'Phones'
);

-- Cleanup
DROP TABLE orders;
DROP TABLE products;
DROP TABLE categories;