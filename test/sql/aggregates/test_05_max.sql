-- Test 5: MAX aggregate
-- Expected: 1 row, max=1500

-- Setup
CREATE TABLE orders (id int, user_id int, product text, amount int);

INSERT INTO orders (id, user_id, product, amount) VALUES
  (1, 1, 'Laptop', 1200),
  (2, 1, 'Mouse', 50),
  (3, 2, 'Keyboard', 100),
  (4, 3, 'Monitor', 300),
  (5, 2, 'Laptop', 1500);

-- Test Query
SELECT MAX(amount) FROM orders;

-- Cleanup
DROP TABLE orders;