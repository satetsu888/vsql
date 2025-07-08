-- Test 27: Subquery in WHERE
-- Expected: 2 rows (Bob, Charlie)

-- Setup
CREATE TABLE users (id int, name text, email text);
INSERT INTO users (id, name, age) VALUES (2, 'Bob', 30);
INSERT INTO users (id, name, email, country) VALUES (3, 'Charlie', 'charlie@example.com', 'USA');

CREATE TABLE orders (id int, user_id int, amount decimal);
INSERT INTO orders VALUES (1, 2, 100.50), (2, 3, 200.00);

-- Test Query
SELECT * FROM users WHERE id IN (SELECT user_id FROM orders);

-- Cleanup
DROP TABLE orders;
DROP TABLE users;