-- Test 24: SUM function
-- Expected: 1 row, sum=300.50

-- Setup
CREATE TABLE orders (id int, user_id int, amount decimal);
INSERT INTO orders VALUES (1, 2, 100.50), (2, 3, 200.00);

-- Test Query
SELECT SUM(amount) FROM orders;

-- Cleanup
DROP TABLE orders;