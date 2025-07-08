-- Test 18: Unicode characters
-- Expected: success

-- Setup
CREATE TABLE users (id int, name text, email text);

-- Test Query
INSERT INTO users (id, name) VALUES (7, '日本語テスト');

-- Cleanup
DROP TABLE users;