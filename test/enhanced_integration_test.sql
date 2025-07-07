-- Enhanced Integration Test Suite
-- Extracted from test_vsql_enhanced.sh

-- ### Basic Table Operations ###

-- Test 1: Create table
-- Expected: success
CREATE TABLE users (id int, name text, email text);

-- Test 2: Insert data
-- Expected: success
INSERT INTO users (id, name, email) VALUES (1, 'Alice', 'alice@example.com');

-- Test 3: Select all
-- Expected: 1 row
SELECT * FROM users;

-- Test 4: Update data
-- Expected: UPDATE 1
UPDATE users SET email = 'alice.new@example.com' WHERE id = 1;

-- Test 5: Delete data
-- Expected: DELETE 1
DELETE FROM users WHERE id = 1;

-- ### Schema-less Features ###

-- Test 6: Insert with new column
-- Expected: success
INSERT INTO users (id, name, age) VALUES (2, 'Bob', 30);

-- Test 7: Select non-existent column (should return NULL)
-- Expected: 1 row with phone=NULL
SELECT id, name, phone FROM users;

-- Test 8: Mixed schema insert
-- Expected: success
INSERT INTO users (id, name, email, country) VALUES (3, 'Charlie', 'charlie@example.com', 'USA');

-- ### Complex WHERE Clauses ###

-- Test 9: AND condition
-- Expected: 1 row (Bob)
SELECT * FROM users WHERE id > 1 AND name = 'Bob';

-- Test 10: OR condition
-- Expected: 2 rows (Bob, Charlie)
SELECT * FROM users WHERE id = 2 OR name = 'Charlie';

-- Test 11: NOT condition
-- Expected: 1 row (Charlie)
SELECT * FROM users WHERE NOT (id = 2);

-- Test 12: Combined conditions
-- Expected: 2 rows
SELECT * FROM users WHERE (id > 1 AND name LIKE 'B%') OR email IS NOT NULL;

-- ### NULL Handling ###

-- Test 13: Insert NULL values
-- Expected: success
INSERT INTO users (id, name, email) VALUES (4, NULL, NULL);

-- Test 14: IS NULL check
-- Expected: 2 rows (Bob with NULL email, id=4)
SELECT * FROM users WHERE email IS NULL;

-- Test 15: IS NOT NULL check
-- Expected: 2 rows (Bob, Charlie)
SELECT * FROM users WHERE name IS NOT NULL;

-- ### Special Characters and Security ###

-- Test 16: Single quotes in data
-- Expected: success
INSERT INTO users (id, name) VALUES (5, 'O''Brien');

-- Test 17: Special characters
-- Expected: success
INSERT INTO users (id, name) VALUES (6, 'Test!@#$%^&*()');

-- Test 18: Unicode characters
-- Expected: success
INSERT INTO users (id, name) VALUES (7, '日本語テスト');

-- ### JOIN Operations ###

-- Test 19: Create orders table
-- Expected: success
CREATE TABLE orders (id int, user_id int, amount decimal);

-- Test 20: Insert order data
-- Expected: success
INSERT INTO orders VALUES (1, 2, 100.50), (2, 3, 200.00);

-- Test 21: Inner join
-- Expected: 2 rows (Bob, Charlie)
SELECT u.name, o.amount FROM users u INNER JOIN orders o ON u.id = o.user_id;

-- Test 22: Left join
-- Expected: 7 rows (all users, some with NULL amount)
SELECT u.name, o.amount FROM users u LEFT JOIN orders o ON u.id = o.user_id;

-- ### Aggregate Functions ###

-- Test 23: COUNT function
-- Expected: 1 row, count=7
SELECT COUNT(*) FROM users;

-- Test 24: SUM function
-- Expected: 1 row, sum=300.50
SELECT SUM(amount) FROM orders;

-- Test 25: AVG function
-- Expected: 1 row, avg=150.25
SELECT AVG(amount) FROM orders;

-- Test 26: GROUP BY
-- Expected: 2 rows (user_id=2: count=1, user_id=3: count=1)
SELECT user_id, COUNT(*) FROM orders GROUP BY user_id;

-- ### Subqueries ###

-- Test 27: Subquery in WHERE
-- Expected: 2 rows (Bob, Charlie)
SELECT * FROM users WHERE id IN (SELECT user_id FROM orders);

-- Test 28: EXISTS subquery
-- Expected: 2 rows (Bob, Charlie)
SELECT * FROM users u WHERE EXISTS (SELECT 1 FROM orders o WHERE o.user_id = u.id);

-- ### Performance Test - Large Dataset ###

-- Test 29: Bulk insert preparation
-- Expected: success (inserting users with id 101-200 for testing)
INSERT INTO users (id, name, email) VALUES 
    (101, 'User101', 'user101@example.com'),
    (102, 'User102', 'user102@example.com'),
    (103, 'User103', 'user103@example.com'),
    (104, 'User104', 'user104@example.com'),
    (105, 'User105', 'user105@example.com'),
    (106, 'User106', 'user106@example.com'),
    (107, 'User107', 'user107@example.com'),
    (108, 'User108', 'user108@example.com'),
    (109, 'User109', 'user109@example.com'),
    (110, 'User110', 'user110@example.com');

-- Test 30: Count after bulk insert
-- Expected: 1 row, count=17
SELECT COUNT(*) FROM users;

-- Test 31: Complex query on larger dataset
-- Expected: 5 rows (users 106-110)
SELECT * FROM users WHERE id > 105 AND email LIKE '%@example.com' ORDER BY id DESC LIMIT 5;

-- ### Cleanup Operations ###

-- Test 32: Drop orders table
-- Expected: success
DROP TABLE orders;

-- Test 33: Drop users table
-- Expected: success
DROP TABLE users;