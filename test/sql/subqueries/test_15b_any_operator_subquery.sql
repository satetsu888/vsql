-- Test: ANY operator with subquery
-- Expected: 2 rows (Laptop, iPhone - products with price > 799.00)
-- Testing ANY operator with subqueries

-- Setup
CREATE TABLE products (id int, name text, category text, price decimal, stock int);

INSERT INTO products VALUES
  (1, 'Laptop', 'Computers', 1200.00, 50),
  (2, 'Mouse', 'Accessories', 25.00, 200),
  (3, 'Keyboard', 'Accessories', 75.00, 150),
  (4, 'iPhone', 'Phones', 999.00, 100),
  (5, 'Android Phone', 'Phones', 799.00, 80);

-- Test Query: ANY operator
SELECT * FROM products
WHERE price > ANY (
  SELECT price FROM products WHERE category = 'Phones'
);

-- Cleanup
DROP TABLE products;