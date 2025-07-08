-- Test 33: Drop users table
-- Expected: success

-- Setup
CREATE TABLE users (id int, name text, email text);

-- Test Query
DROP TABLE users;