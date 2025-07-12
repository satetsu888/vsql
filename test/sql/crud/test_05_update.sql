-- Test 4: Update data
-- Expected: success

-- Setup
CREATE TABLE users (id int, name text, email text);
INSERT INTO users (id, name, email) VALUES (1, 'Alice', 'alice@example.com');

-- Test Query
UPDATE users SET email = 'alice.new@example.com' WHERE id = 1;

-- Cleanup
DROP TABLE users;