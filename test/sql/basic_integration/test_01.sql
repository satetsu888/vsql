-- Test 1: Create table with schema-less design
-- Expected: success

-- Test Query
CREATE TABLE users (id int, name text, email text);

-- Cleanup
DROP TABLE users;