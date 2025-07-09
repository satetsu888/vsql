-- Test 9: GROUP BY with HAVING
-- Expected: 1 row (Tokyo: 2)

-- Setup
CREATE TABLE users (id int, name text, age int, city text);

INSERT INTO users (id, name, age, city) VALUES 
  (1, 'Alice', 25, 'Tokyo'),
  (2, 'Bob', 30, 'Osaka'),
  (3, 'Charlie', 35, 'Tokyo'),
  (4, 'David', 28, 'Kyoto');

-- Test Query
SELECT city, COUNT(*) as count 
FROM users 
GROUP BY city 
HAVING COUNT(*) > 1;

-- Cleanup
DROP TABLE users;