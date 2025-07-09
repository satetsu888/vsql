-- Test 2: ORDER BY DESC
-- Expected: 4 rows ordered by age descending (Charlie: 35, Bob: 30, David: 28, Alice: 25)
-- Based on: basic_advanced/test_12.sql

-- Setup
CREATE TABLE users (id int, name text, age int, city text);

INSERT INTO users (id, name, age, city) VALUES 
  (1, 'Alice', 25, 'Tokyo'),
  (2, 'Bob', 30, 'Osaka'),
  (3, 'Charlie', 35, 'Tokyo'),
  (4, 'David', 28, 'Kyoto');

-- Test Query
SELECT * FROM users ORDER BY age DESC;

-- Cleanup
DROP TABLE users;