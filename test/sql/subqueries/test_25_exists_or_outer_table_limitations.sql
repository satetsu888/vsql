-- Test: Complex OR conditions with outer table references in EXISTS
-- Status: Expected to FAIL - OR conditions referencing outer tables not fully supported
-- These tests document the limitation with OR conditions that reference outer query tables

-- Setup
CREATE TABLE users (id int, name text, active int, country text);
CREATE TABLE posts (id int, user_id int, title text);
CREATE TABLE comments (id int, post_id int, user_id int, content text);

INSERT INTO users VALUES
  (1, 'Alice', 1, 'Japan'),
  (2, 'Bob', 1, 'USA'),
  (3, 'Charlie', 0, 'Japan'),
  (4, 'David', 0, 'USA'),
  (5, 'Eve', 1, 'UK');

INSERT INTO posts VALUES
  (1, 1, 'Post by Alice'),
  (2, 3, 'Post by Charlie'),
  (3, 5, 'Post by Eve');

INSERT INTO comments VALUES
  (1, 1, 2, 'Comment by Bob'),
  (2, 2, 1, 'Comment by Alice');

-- Test 1: Simple OR with outer table column reference
-- Expected: Should return all active users (Alice, Bob, Eve)
-- Actual: Only returns users with posts (Alice)
-- Issue: The OR condition "u.active = true" is not evaluated correctly
SELECT name FROM users u
WHERE EXISTS (
  SELECT 1 FROM posts p 
  WHERE p.user_id = u.id 
  OR u.active = 1
)
ORDER BY name;

-- Test 2: OR condition mixing inner and outer table columns
-- Expected: Should return users who have posts OR are from Japan (Alice, Charlie, Eve)
-- Actual: Only returns users with posts (Alice, Charlie, Eve) - might work by coincidence
SELECT name FROM users u
WHERE EXISTS (
  SELECT 1 FROM posts p 
  WHERE p.user_id = u.id 
  OR u.country = 'Japan'
)
ORDER BY name;

-- Test 3: Complex OR with multiple outer references
-- Expected: Should return active users OR users from USA (Alice, Bob, David, Eve)
-- Actual: Only returns users with posts
SELECT name FROM users u
WHERE EXISTS (
  SELECT 1 FROM posts p 
  WHERE p.user_id = u.id 
  OR (u.active = 1 OR u.country = 'USA')
)
ORDER BY name;

-- Test 4: NOT EXISTS with OR referencing outer table
-- Expected: Should return users without posts AND not active (Charlie, David)
-- Actual: May return incorrect results
SELECT name FROM users u
WHERE NOT EXISTS (
  SELECT 1 FROM posts p 
  WHERE p.user_id = u.id 
  OR u.active = 1
)
ORDER BY name;

-- Test 5: Nested EXISTS with OR conditions
-- Expected: Should return users who have posts with comments OR are active
-- Actual: Only returns users with posts that have comments
SELECT name FROM users u
WHERE EXISTS (
  SELECT 1 FROM posts p 
  WHERE p.user_id = u.id
  AND EXISTS (
    SELECT 1 FROM comments c 
    WHERE c.post_id = p.id 
    OR u.active = 1
  )
)
ORDER BY name;

-- Test 6: OR with function on outer table column
-- Expected: Should return users with posts OR users with name starting with 'B'
-- Actual: Only returns users with posts
SELECT name FROM users u
WHERE EXISTS (
  SELECT 1 FROM posts p 
  WHERE p.user_id = u.id 
  OR u.name LIKE 'B%'
)
ORDER BY name;

-- Cleanup
DROP TABLE comments;
DROP TABLE posts;
DROP TABLE users;