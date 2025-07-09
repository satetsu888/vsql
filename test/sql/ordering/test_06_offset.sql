-- Test 6: OFFSET without LIMIT (rarely used but valid)
-- Expected: Skip first 2 rows, return remaining rows

-- Setup
CREATE TABLE users (id int, name text, age int, city text);

INSERT INTO users (id, name, age, city) VALUES 
  (1, 'Alice', 25, 'Tokyo'),
  (2, 'Bob', 30, 'Osaka'),
  (3, 'Charlie', 35, 'Tokyo'),
  (4, 'David', 28, 'Kyoto'),
  (5, 'Eve', 32, 'Osaka');

-- Test Query
SELECT * FROM users ORDER BY id OFFSET 2;

-- Cleanup
DROP TABLE users;