-- Test 7: GROUP BY city
-- Expected: 3 rows (Tokyo: 2, Osaka: 1, Kyoto: 1)

-- Setup
CREATE TABLE users (id int, name text, age int, city text);

INSERT INTO users (id, name, age, city) VALUES 
  (1, 'Alice', 25, 'Tokyo'),
  (2, 'Bob', 30, 'Osaka'),
  (3, 'Charlie', 35, 'Tokyo'),
  (4, 'David', 28, 'Kyoto');

-- Test Query
SELECT city, COUNT(*) as user_count 
FROM users 
GROUP BY city
ORDER BY city;

-- Cleanup
DROP TABLE users;