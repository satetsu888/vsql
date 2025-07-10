-- Test 10: Multiple JOINs (three tables)
-- Expected: 6 rows
-- Status: PASS
-- Test: SELECT u.name, o.amount, p.name as product_name, p.price FROM users u JOIN orders o ON u.id = o.user_id JOIN products p ON o.product = p.name ORDER BY u.name, o.amount

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

CREATE TABLE products (
    name TEXT,
    price INTEGER,
    category TEXT
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

INSERT INTO products (name, price, category) VALUES
    ('Laptop', 1500, 'Electronics'),
    ('Mouse', 50, 'Accessories'),
    ('Keyboard', 100, 'Accessories'),
    ('Monitor', 800, 'Electronics');

-- Test query
SELECT u.name, o.amount, p.name as product_name, p.price
FROM users u
JOIN orders o ON u.id = o.user_id
JOIN products p ON o.product = p.name
ORDER BY u.name, o.amount;

-- Cleanup
DROP TABLE products;
DROP TABLE orders;
DROP TABLE users;