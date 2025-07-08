-- Test 12: Complex WHERE with BETWEEN and IN
-- Expected: 5 rows (all users match the conditions)
-- Fixed: All users match - Alice(30,Tokyo), Bob(25,Osaka), Charlie(35,Tokyo), Dave(45,Kyoto), Eve(28,Osaka)
-- Test: SELECT * FROM users WHERE (age BETWEEN 25 AND 35 AND city IN ('Tokyo', 'Osaka')) OR (age > 40 AND city = 'Kyoto') ORDER BY age DESC, name ASC

-- Create test tables
CREATE TABLE users (
    id INTEGER,
    name TEXT,
    age INTEGER,
    city TEXT
);

-- Insert test data
INSERT INTO users (id, name, age, city) VALUES
    (1, 'Alice', 30, 'Tokyo'),
    (2, 'Bob', 25, 'Osaka'),
    (3, 'Charlie', 35, 'Tokyo'),
    (4, 'Dave', 45, 'Kyoto'),
    (5, 'Eve', 28, 'Osaka');

-- Test query
SELECT * FROM users
WHERE (age BETWEEN 25 AND 35 AND city IN ('Tokyo', 'Osaka'))
   OR (age > 40 AND city = 'Kyoto')
ORDER BY age DESC, name ASC;

-- Cleanup
DROP TABLE users;