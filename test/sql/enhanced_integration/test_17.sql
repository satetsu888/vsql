-- Test 17: Special characters
-- Expected: success

-- Setup
CREATE TABLE users (id int, name text, email text);

-- Test Query
INSERT INTO users (id, name) VALUES (6, 'Test!@#$%^&*()');

-- Cleanup
DROP TABLE users;