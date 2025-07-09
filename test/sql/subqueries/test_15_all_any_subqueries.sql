-- Test 15: ALL/ANY subqueries
-- Testing ALL and ANY operators with subqueries
-- Status: May fail - ALL/ANY operators not implemented

-- Setup
CREATE TABLE products (id int, name text, category text, price decimal, stock int);

INSERT INTO products VALUES
  (1, 'Laptop', 'Computers', 1200.00, 50),
  (2, 'Mouse', 'Accessories', 25.00, 200),
  (3, 'Keyboard', 'Accessories', 75.00, 150),
  (4, 'iPhone', 'Phones', 999.00, 100),
  (5, 'Android Phone', 'Phones', 799.00, 80);

-- Test Query 1: ALL operator
SELECT * FROM products
WHERE price > ALL (
  SELECT price FROM products WHERE category = 'Accessories'
);

-- Test Query 2: ANY operator
SELECT * FROM products
WHERE price > ANY (
  SELECT price FROM products WHERE category = 'Phones'
);

-- Cleanup
DROP TABLE products;