-- Test 3: ORDER BY multiple columns
-- Expected: 6 rows ordered first by city, then by name within each city
-- Based on: basic_advanced/test_07.sql

-- Setup
CREATE TABLE users (id int, name text, age int, city text);

INSERT INTO users (id, name, age, city) VALUES 
  (1, 'Alice', 25, 'Tokyo'),
  (2, 'Bob', 30, 'Osaka'),
  (3, 'Charlie', 35, 'Tokyo'),
  (4, 'David', 28, 'Kyoto'),
  (5, 'Eve', 32, 'Osaka'),
  (6, 'Frank', 29, 'Tokyo');

-- Test Query
SELECT city, name, age FROM users ORDER BY city, name;

-- Cleanup
DROP TABLE users;