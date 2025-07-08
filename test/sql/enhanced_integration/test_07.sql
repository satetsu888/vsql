-- Test 7: Select non-existent column (should return NULL)
-- Expected: 1 row with phone=NULL

-- Setup
CREATE TABLE users (id int, name text, email text);
INSERT INTO users (id, name, age) VALUES (2, 'Bob', 30);

-- Test Query
SELECT id, name, phone FROM users;

-- Cleanup
DROP TABLE users;