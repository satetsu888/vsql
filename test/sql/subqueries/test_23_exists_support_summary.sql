-- EXISTS/NOT EXISTS Support Summary
-- This file documents what works and what doesn't with EXISTS/NOT EXISTS in VSQL

-- Setup
CREATE TABLE users (id int, name text, active boolean);
CREATE TABLE posts (id int, user_id int, title text);

INSERT INTO users VALUES
  (1, 'Alice', true),
  (2, 'Bob', true),
  (3, 'Charlie', false);

INSERT INTO posts VALUES
  (1, 1, 'First post'),
  (2, 1, 'Second post'),
  (3, 2, 'Bob post');

-- WORKING FEATURES:

-- 1. Basic correlated EXISTS ✓
-- Expected: 2 rows (Alice, Bob)
SELECT name FROM users u
WHERE EXISTS (SELECT 1 FROM posts p WHERE p.user_id = u.id)
ORDER BY name;

-- 2. Basic correlated NOT EXISTS ✓
-- Expected: 1 row (Charlie)
SELECT name FROM users u
WHERE NOT EXISTS (SELECT 1 FROM posts p WHERE p.user_id = u.id);

-- 3. Non-correlated EXISTS ✓
-- Expected: 3 rows (all users, because posts table has rows)
SELECT name FROM users
WHERE EXISTS (SELECT 1 FROM posts)
ORDER BY name;

-- 4. EXISTS with additional WHERE conditions ✓
-- Expected: 1 row (Alice)
SELECT name FROM users u
WHERE EXISTS (SELECT 1 FROM posts p WHERE p.user_id = u.id AND p.title LIKE '%First%');

-- 5. EXISTS with calculations in correlation ✓
-- Expected: 2 rows (Alice, Bob - users with id < 3 who have posts)
SELECT name FROM users u
WHERE u.id < 3
AND EXISTS (SELECT 1 FROM posts p WHERE p.user_id = u.id)
ORDER BY name;

-- NOT WORKING FEATURES:

-- 6. Nested EXISTS ✗
-- Status: Returns 0 rows instead of expected results
-- This query attempts to use EXISTS inside another EXISTS
SELECT name FROM users u
WHERE EXISTS (
  SELECT 1 FROM posts p 
  WHERE p.user_id = u.id
  AND EXISTS (SELECT 1 FROM users u2 WHERE u2.id = p.user_id AND u2.active = true)
);

-- 7. EXISTS with GROUP BY/HAVING ✗
-- Status: Returns 0 rows instead of expected results
-- This query attempts to use aggregate functions in EXISTS subquery
SELECT name FROM users u
WHERE EXISTS (
  SELECT 1 FROM posts p 
  WHERE p.user_id = u.id
  GROUP BY p.user_id
  HAVING COUNT(*) > 1
);

-- 8. EXISTS with OR referencing outer table ✗
-- Status: OR condition referencing outer table is ignored
-- Expected: All active users (even without posts)
-- Actual: Only users with posts
SELECT name FROM users u
WHERE EXISTS (
  SELECT 1 FROM posts p 
  WHERE p.user_id = u.id 
  OR u.active = true
)
ORDER BY name;

-- SUMMARY:
-- ✓ Basic correlated EXISTS/NOT EXISTS work well
-- ✓ Simple WHERE conditions in EXISTS subqueries work
-- ✓ Non-correlated EXISTS works
-- ✗ Nested EXISTS not supported
-- ✗ GROUP BY/HAVING in EXISTS not supported  
-- ✗ Complex OR conditions referencing outer table may not work correctly

-- Cleanup
DROP TABLE posts;
DROP TABLE users;