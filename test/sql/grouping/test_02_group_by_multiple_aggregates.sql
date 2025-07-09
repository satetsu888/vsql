-- Test 6: GROUP BY with COUNT and AVG
-- Expected: 3 rows (Tokyo: 2 users, avg_age=32.5; Osaka: 2 users, avg_age=26.5; Kyoto: 1 user, avg_age=45)
-- Test: SELECT city, COUNT(*) as user_count, AVG(age) as avg_age FROM users GROUP BY city ORDER BY city

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
SELECT city, COUNT(*) as user_count, AVG(age) as avg_age
FROM users 
GROUP BY city
ORDER BY city;

-- Cleanup
DROP TABLE users;