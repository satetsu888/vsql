-- Test: Comprehensive EXISTS/NOT EXISTS test cases
-- This file tests various edge cases and complex scenarios

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

-- Test 1: EXISTS with multiple correlation conditions
-- Expected: 3 rows (Laptop, Mouse, Desk - products with orders)
SELECT name FROM products p
WHERE EXISTS (
  SELECT 1 FROM orders o 
  WHERE o.product_id = p.id
)
ORDER BY name;

-- Test 2: NOT EXISTS to find products without inventory
-- Expected: 2 rows (Pen, Notebook)
SELECT name FROM products p
WHERE NOT EXISTS (
  SELECT 1 FROM inventory i 
  WHERE i.product_id = p.id
)
ORDER BY name;

-- Test 3: Nested EXISTS 
-- Status: FAILS - Nested EXISTS not supported
-- Expected: 3 rows (Laptop, Mouse, Desk - products with both inventory and orders)
-- Actual: 0 rows
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

-- Test 4: EXISTS with aggregate subquery
-- Status: FAILS - EXISTS with GROUP BY/HAVING not supported  
-- Expected: 2 rows (Laptop, Desk - products with total inventory > 15)
-- Actual: 0 rows
SELECT name FROM products p
WHERE EXISTS (
  SELECT 1 FROM inventory i 
  WHERE i.product_id = p.id
  GROUP BY i.product_id
  HAVING SUM(i.quantity) > 15
)
ORDER BY name;

-- Test 5: NOT EXISTS with JOIN in subquery
-- Expected: 2 rows (Pen, Notebook - products with no inventory in any warehouse)
SELECT name FROM products p
WHERE NOT EXISTS (
  SELECT 1 
  FROM inventory i 
  INNER JOIN products p2 ON i.product_id = p2.id
  WHERE p2.id = p.id
)
ORDER BY name;

-- Test 6: EXISTS with OR conditions in correlation
-- Status: PARTIALLY WORKING - Returns only products with orders, ignores OR condition
-- Expected: 4 rows (all Electronics and products with orders: Laptop, Mouse, Desk, Chair)
-- Actual: 3 rows (only products with orders: Laptop, Mouse, Desk)
SELECT name FROM products p
WHERE EXISTS (
  SELECT 1 FROM orders o 
  WHERE o.product_id = p.id 
  OR p.category = 'Electronics'
)
ORDER BY name;

-- Test 7: Complex correlated EXISTS with calculations
-- Expected: 2 rows (Laptop, Desk - expensive products with orders)
SELECT name FROM products p
WHERE p.price > 100
AND EXISTS (
  SELECT 1 FROM orders o 
  WHERE o.product_id = p.id 
  AND o.quantity * p.price > 500
)
ORDER BY name;

-- Test 8: NOT EXISTS with NULL handling
-- Add some NULL data
INSERT INTO inventory VALUES (NULL, 'Tokyo', 50);
INSERT INTO orders VALUES (5, NULL, 'Dave', 2);

-- Expected: Still 2 rows (Pen, Notebook - NULL product_id doesn't match)
SELECT name FROM products p
WHERE NOT EXISTS (
  SELECT 1 FROM inventory i 
  WHERE i.product_id = p.id
)
ORDER BY name;

-- Cleanup
DROP TABLE orders;
DROP TABLE inventory;
DROP TABLE products;