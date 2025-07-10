-- Test 14: JOIN with table-qualified columns
-- Expected: 2 rows
-- Test: SELECT u.name, o.amount FROM users u JOIN orders o ON u.id = o.user_id

-- Create test tables
CREATE TABLE users (
    id INTEGER,
    name TEXT
);

CREATE TABLE orders (
    id INTEGER,
    user_id INTEGER,
    amount INTEGER
);

-- Insert test data
INSERT INTO users (id, name) VALUES
    (1, 'Alice'),
    (2, 'Bob');

INSERT INTO orders (id, user_id, amount) VALUES
    (101, 1, 100),
    (102, 2, 200);

-- Test query with table-qualified columns
SELECT u.name, o.amount 
FROM users u 
JOIN orders o ON u.id = o.user_id;

-- Test query with both tables having same column name
-- Expected: 2 rows
SELECT u.id, o.id, u.name 
FROM users u 
JOIN orders o ON u.id = o.user_id;

-- Cleanup
DROP TABLE orders;
DROP TABLE users;