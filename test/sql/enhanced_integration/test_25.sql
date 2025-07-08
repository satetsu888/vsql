-- Test 25: AVG function
-- Expected: 1 row, avg=150.25

-- Setup
CREATE TABLE orders (id int, user_id int, amount decimal);
INSERT INTO orders VALUES (1, 2, 100.50), (2, 3, 200.00);

-- Test Query
SELECT AVG(amount) FROM orders;

-- Cleanup
DROP TABLE orders;