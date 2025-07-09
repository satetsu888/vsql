-- Test 4: ORDER BY with NULL values
-- Expected: 5 rows with NULL ages appearing last (or first depending on implementation)

-- Setup
CREATE TABLE users (id int, name text, age int, city text);

INSERT INTO users (id, name, age, city) VALUES 
  (1, 'Alice', 25, 'Tokyo'),
  (2, 'Bob', NULL, 'Osaka'),
  (3, 'Charlie', 35, 'Tokyo'),
  (4, 'David', 28, 'Kyoto'),
  (5, 'Eve', NULL, 'Osaka');

-- Test Query
SELECT name, age FROM users ORDER BY age;

-- Cleanup
DROP TABLE users;