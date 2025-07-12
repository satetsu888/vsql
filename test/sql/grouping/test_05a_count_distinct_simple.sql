-- Test: COUNT DISTINCT function - Simple case
-- Expected: 1 row
-- Test: Verifies COUNT DISTINCT works correctly with single table

-- Setup
CREATE TABLE users (id int, name text, age int, city text);

INSERT INTO users (id, name, age, city) VALUES 
  (1, 'Alice', 25, 'Tokyo'),
  (2, 'Bob', 30, 'Osaka'),
  (3, 'Charlie', 35, 'Tokyo'),
  (4, 'David', 28, 'Kyoto');

-- Test Query - Simple COUNT DISTINCT
SELECT COUNT(DISTINCT u.city) as unique_cities FROM users u;

-- Cleanup
DROP TABLE users;