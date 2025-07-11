-- Test: NOT EXISTS with JOIN in subquery
-- Expected: 2 rows (Pen, Notebook - products with no inventory in any warehouse)

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
WHERE NOT EXISTS (
  SELECT 1 
  FROM inventory i 
  INNER JOIN products p2 ON i.product_id = p2.id
  WHERE p2.id = p.id
)
ORDER BY name;

-- Cleanup
DROP TABLE inventory;
DROP TABLE products;