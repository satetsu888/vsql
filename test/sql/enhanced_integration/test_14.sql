-- Test 14: IS NULL check
-- Expected: 5 rows (Bob, id=4, O'Brien, Test!@#$%^&*(), 日本語テスト - all have NULL email)
-- Note: This test assumes additional data from later tests

-- Setup
CREATE TABLE users (id int, name text, email text);
INSERT INTO users (id, name, email) VALUES (1, 'Alice', 'alice@example.com');
INSERT INTO users (id, name, age) VALUES (2, 'Bob', 30);
INSERT INTO users (id, name, email, country) VALUES (3, 'Charlie', 'charlie@example.com', 'USA');
INSERT INTO users (id, name, email) VALUES (4, NULL, NULL);
INSERT INTO users (id, name) VALUES (5, 'O''Brien');
INSERT INTO users (id, name) VALUES (6, 'Test!@#$%^&*()');
INSERT INTO users (id, name) VALUES (7, '日本語テスト');

-- Test Query
SELECT * FROM users WHERE email IS NULL;

-- Cleanup
DROP TABLE users;