-- Test 20: Insert order data
-- Expected: success

-- Setup
CREATE TABLE orders (id int, user_id int, amount decimal);

-- Test Query
INSERT INTO orders VALUES (1, 2, 100.50), (2, 3, 200.00);

-- Cleanup
DROP TABLE orders;