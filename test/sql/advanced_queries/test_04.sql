-- Test 4: EXISTS subquery
-- Expected: 2 rows (Alice, Charlie)
-- Fixed: Dave only has amount=45 (< 500) so shouldn't be included
-- Test: SELECT name FROM users u WHERE EXISTS (SELECT 1 FROM orders o WHERE o.user_id = u.id AND o.amount > 500) ORDER BY name

-- Create test tables
CREATE TABLE users (
    id INTEGER,
    name TEXT,
    age INTEGER,
    city TEXT
);

CREATE TABLE orders (
    id INTEGER,
    user_id INTEGER,
    product TEXT,
    amount INTEGER
);

-- Insert test data
INSERT INTO users (id, name, age, city) VALUES
    (1, 'Alice', 30, 'Tokyo'),
    (2, 'Bob', 25, 'Osaka'),
    (3, 'Charlie', 35, 'Tokyo'),
    (4, 'Dave', 45, 'Kyoto'),
    (5, 'Eve', 28, 'Osaka');

INSERT INTO orders (id, user_id, product, amount) VALUES
    (1, 1, 'Laptop', 1500),
    (2, 1, 'Mouse', 50),
    (3, 2, 'Keyboard', 100),
    (4, 3, 'Monitor', 800),
    (5, 3, 'Laptop', 1600),
    (6, 4, 'Mouse', 45);

-- Test query
SELECT name FROM users u
WHERE EXISTS (SELECT 1 FROM orders o WHERE o.user_id = u.id AND o.amount > 500)
ORDER BY name;

-- Cleanup
DROP TABLE orders;
DROP TABLE users;