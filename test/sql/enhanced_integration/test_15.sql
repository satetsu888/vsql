-- Test 15: IS NOT NULL check
-- Expected: 6 rows (all users except id=4 have non-NULL names)
-- Note: Updated to match actual test data

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
SELECT * FROM users WHERE name IS NOT NULL;

-- Cleanup
DROP TABLE users;