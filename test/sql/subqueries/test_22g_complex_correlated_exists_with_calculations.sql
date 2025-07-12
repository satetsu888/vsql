-- Test: Complex correlated EXISTS with calculations
-- Expected: 2 rows (Laptop, Desk - expensive products with orders)

-- Setup
CREATE TABLE products (id int, name text, category text, price int);
CREATE TABLE orders (id int, product_id int, customer text, quantity int);

INSERT INTO products VALUES
  (1, 'Laptop', 'Electronics', 1000),
  (2, 'Mouse', 'Electronics', 50),
  (3, 'Desk', 'Furniture', 300),
  (4, 'Chair', 'Furniture', 200),
  (5, 'Pen', 'Stationery', 5),
  (6, 'Notebook', 'Stationery', 10);

INSERT INTO orders VALUES
  (1, 1, 'Alice', 2),
  (2, 1, 'Bob', 1),
  (3, 2, 'Alice', 5),
  (4, 3, 'Charlie', 1);
-- Note: Products 4, 5, 6 have no orders

-- Test query
SELECT name FROM products p
WHERE p.price > 100
AND EXISTS (
  SELECT 1 FROM orders o 
  WHERE o.product_id = p.id 
  AND o.quantity * p.price >= 300
)
ORDER BY name;

-- Cleanup
DROP TABLE orders;
DROP TABLE products;