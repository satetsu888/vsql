-- Test file to verify JOIN with table-qualified column names

-- Create tables
CREATE TABLE users (id int, name text);
CREATE TABLE orders (id int, user_id int, amount int);

-- Insert test data
INSERT INTO users VALUES (1, 'Alice'), (2, 'Bob');
INSERT INTO orders VALUES (101, 1, 100), (102, 2, 200);

-- Test 1: Basic JOIN without qualified columns
-- Expected: 2
SELECT name, amount FROM users u JOIN orders o ON u.id = o.user_id;

-- Test 2: JOIN with table-qualified columns
-- Expected: 2
SELECT u.name, o.amount FROM users u JOIN orders o ON u.id = o.user_id;

-- Test 3: JOIN with both tables having same column name
-- Expected: 2
SELECT u.id, o.id, u.name FROM users u JOIN orders o ON u.id = o.user_id;

-- Clean up
DROP TABLE users;
DROP TABLE orders;