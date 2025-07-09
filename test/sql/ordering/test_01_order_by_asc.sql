-- Test 1: Basic ORDER BY ASC
-- Expected: 4 rows ordered by name alphabetically (Alice, Bob, Charlie, David)

-- Setup
CREATE TABLE users (id int, name text, age int, city text);

INSERT INTO users (id, name, age, city) VALUES 
  (1, 'Alice', 25, 'Tokyo'),
  (2, 'Bob', 30, 'Osaka'),
  (3, 'Charlie', 35, 'Tokyo'),
  (4, 'David', 28, 'Kyoto');

-- Test Query
SELECT * FROM users ORDER BY name;

-- Cleanup
DROP TABLE users;