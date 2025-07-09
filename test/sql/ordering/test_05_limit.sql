-- Test 5: Basic LIMIT
-- Expected: 2 rows (Charlie: 35, Bob: 30)
-- Based on: basic_advanced/test_12.sql

-- Setup
CREATE TABLE users (id int, name text, age int, city text);

INSERT INTO users (id, name, age, city) VALUES 
  (1, 'Alice', 25, 'Tokyo'),
  (2, 'Bob', 30, 'Osaka'),
  (3, 'Charlie', 35, 'Tokyo'),
  (4, 'David', 28, 'Kyoto');

-- Test Query
SELECT * FROM users ORDER BY age DESC LIMIT 2;

-- Cleanup
DROP TABLE users;