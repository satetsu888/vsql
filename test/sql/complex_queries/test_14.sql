-- Test 14: Complex CASE statements
-- Testing CASE WHEN expressions
-- Status: May fail - CASE expressions not implemented

-- Setup
CREATE TABLE products (id int, name text, category text, price decimal, stock int);

INSERT INTO products VALUES
  (1, 'Laptop', 'Computers', 1200.00, 50),
  (2, 'Mouse', 'Accessories', 25.00, 200),
  (3, 'Keyboard', 'Accessories', 75.00, 150),
  (4, 'iPhone', 'Phones', 999.00, 100),
  (5, 'Android Phone', 'Phones', 799.00, 80);

-- Test Query
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

-- Cleanup
DROP TABLE products;