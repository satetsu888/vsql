-- Test 7: LIMIT with OFFSET (pagination)
-- Expected: 2 rows (rows 3-4 when ordered by id)

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
SELECT * FROM users ORDER BY id LIMIT 2 OFFSET 2;

-- Cleanup
DROP TABLE users;