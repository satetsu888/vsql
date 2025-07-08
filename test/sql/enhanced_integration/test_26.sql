-- Test 26: GROUP BY
-- Expected: 2 rows (user_id=2: count=1, user_id=3: count=1)

-- Setup
CREATE TABLE orders (id int, user_id int, amount decimal);
INSERT INTO orders VALUES (1, 2, 100.50), (2, 3, 200.00);

-- Test Query
SELECT user_id, COUNT(*) FROM orders GROUP BY user_id;

-- Cleanup
DROP TABLE orders;