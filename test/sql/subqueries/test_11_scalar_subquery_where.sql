-- Test 11: Scalar subquery comparison
-- Expected: 2 rows (Bob: 30, Charlie: 35)

-- Setup
CREATE TABLE users (id int, name text, age int, city text);

INSERT INTO users (id, name, age, city) VALUES 
  (1, 'Alice', 25, 'Tokyo'),
  (2, 'Bob', 30, 'Osaka'),
  (3, 'Charlie', 35, 'Tokyo'),
  (4, 'David', 28, 'Kyoto');

-- Test Query
SELECT name, age 
FROM users 
WHERE age > (SELECT AVG(age) FROM users)
ORDER BY name;

-- Cleanup
DROP TABLE users;