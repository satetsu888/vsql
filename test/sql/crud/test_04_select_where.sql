-- Test 10: OR condition
-- Expected: 2 rows (Bob, Charlie)

-- Setup
CREATE TABLE users (id int, name text, email text);
INSERT INTO users (id, name, age) VALUES (2, 'Bob', 30);
INSERT INTO users (id, name, email, country) VALUES (3, 'Charlie', 'charlie@example.com', 'USA');

-- Test Query
SELECT * FROM users WHERE id = 2 OR name = 'Charlie';

-- Cleanup
DROP TABLE users;