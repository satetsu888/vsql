-- Test 12: Combined conditions
-- Expected: 3 rows (Bob matches first condition, plus Alice and Charlie have email)
-- Note: Updated to match actual test data

-- Setup
CREATE TABLE users (id int, name text, email text);
INSERT INTO users (id, name, email) VALUES (1, 'Alice', 'alice@example.com');
INSERT INTO users (id, name, age) VALUES (2, 'Bob', 30);
INSERT INTO users (id, name, email, country) VALUES (3, 'Charlie', 'charlie@example.com', 'USA');

-- Test Query
SELECT * FROM users WHERE (id > 1 AND name LIKE 'B%') OR email IS NOT NULL;

-- Cleanup
DROP TABLE users;