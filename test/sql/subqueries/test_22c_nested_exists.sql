-- Test: Nested EXISTS 
-- Status: FAILS - Nested EXISTS not supported
-- Expected: 3 rows (Laptop, Mouse, Desk - products with both inventory and orders)
-- Actual: 0 rows

-- Setup
CREATE TABLE products (id int, name text, category text, price int);
CREATE TABLE inventory (product_id int, warehouse text, quantity int);
CREATE TABLE orders (id int, product_id int, customer text, quantity int);

INSERT INTO products VALUES
  (1, 'Laptop', 'Electronics', 1000),
  (2, 'Mouse', 'Electronics', 50),
  (3, 'Desk', 'Furniture', 300),
  (4, 'Chair', 'Furniture', 200),
  (5, 'Pen', 'Stationery', 5),
  (6, 'Notebook', 'Stationery', 10);

INSERT INTO inventory VALUES
  (1, 'Tokyo', 10),
  (1, 'Osaka', 5),
  (2, 'Tokyo', 100),
  (3, 'Tokyo', 20),
  (4, 'Osaka', 15);
-- Note: Products 5 and 6 have no inventory

INSERT INTO orders VALUES
  (1, 1, 'Alice', 2),
  (2, 1, 'Bob', 1),
  (3, 2, 'Alice', 5),
  (4, 3, 'Charlie', 1);
-- Note: Products 4, 5, 6 have no orders

-- Test query
SELECT name FROM products p
WHERE EXISTS (
  SELECT 1 FROM inventory i 
  WHERE i.product_id = p.id
  AND EXISTS (
    SELECT 1 FROM orders o 
    WHERE o.product_id = p.id
  )
)
ORDER BY name;

-- Cleanup
DROP TABLE orders;
DROP TABLE inventory;
DROP TABLE products;