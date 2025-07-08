-- Test 13: Insert NULL values
-- Expected: success

-- Setup
CREATE TABLE users (id int, name text, email text);

-- Test Query
INSERT INTO users (id, name, email) VALUES (4, NULL, NULL);

-- Cleanup
DROP TABLE users;