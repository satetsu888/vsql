-- Test: EXISTS with aggregate subquery
-- Status: FAILS - EXISTS with GROUP BY/HAVING not supported  
-- Expected: 2 rows (Laptop, Desk - products with total inventory > 15)
-- Actual: 0 rows

-- Setup
CREATE TABLE products (id int, name text, category text, price int);
CREATE TABLE inventory (product_id int, warehouse text, quantity int);

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

-- Test query
SELECT name FROM products p
WHERE EXISTS (
  SELECT 1 FROM inventory i 
  WHERE i.product_id = p.id
  GROUP BY i.product_id
  HAVING SUM(i.quantity) > 15
)
ORDER BY name;

-- Cleanup
DROP TABLE inventory;
DROP TABLE products;