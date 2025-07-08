-- Test 19: Create orders table
-- Expected: success

-- Test Query
CREATE TABLE orders (id int, user_id int, amount decimal);

-- Cleanup
DROP TABLE orders;