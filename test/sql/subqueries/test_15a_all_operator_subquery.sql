-- Test: ALL operator with subquery
-- Expected: 3 rows (Laptop, iPhone, Android Phone - products with price > 75.00)
-- Status: FAILING - ALL operator not implemented
-- Testing ALL operator with subqueries

-- Setup
CREATE TABLE products (id int, name text, category text, price decimal, stock int);

INSERT INTO products VALUES
  (1, 'Laptop', 'Computers', 1200.00, 50),
  (2, 'Mouse', 'Accessories', 25.00, 200),
  (3, 'Keyboard', 'Accessories', 75.00, 150),
  (4, 'iPhone', 'Phones', 999.00, 100),
  (5, 'Android Phone', 'Phones', 799.00, 80);

-- Test Query: ALL operator
SELECT * FROM products
WHERE price > ALL (
  SELECT price FROM products WHERE category = 'Accessories'
);

-- Cleanup
DROP TABLE products;