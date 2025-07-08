-- Test 32: Drop orders table
-- Expected: success

-- Setup
CREATE TABLE orders (id int, user_id int, amount decimal);

-- Test Query
DROP TABLE orders;