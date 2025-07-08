-- Test 12: Window functions (if supported)
-- Testing window functions OVER PARTITION BY
-- Status: May fail - window functions not implemented

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
  user_id,
  product_id,
  quantity,
  price,
  SUM(price * quantity) OVER (PARTITION BY user_id) as user_total,
  ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY price DESC) as price_rank
FROM orders;

-- Cleanup
DROP TABLE orders;
DROP TABLE products;