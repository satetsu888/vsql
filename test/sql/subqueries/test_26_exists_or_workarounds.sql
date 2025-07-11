-- Test: Workarounds for OR conditions with outer table references
-- This file shows alternative query patterns that work correctly
-- FAILING: This test file contains features that are not yet fully implemented

-- Setup
CREATE TABLE users (id int, name text, active int, country text);
CREATE TABLE posts (id int, user_id int, title text);

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

-- Workaround 1: Use UNION instead of OR
-- Original (doesn't work):
-- SELECT name FROM users u WHERE EXISTS (SELECT 1 FROM posts p WHERE p.user_id = u.id OR u.active = 1)
-- Workaround (works):
-- Expected: 5 rows (all active users + users with posts)
SELECT name FROM users u WHERE u.active = 1
UNION
SELECT name FROM users u WHERE EXISTS (SELECT 1 FROM posts p WHERE p.user_id = u.id)
ORDER BY name;

-- Workaround 2: Move outer table condition outside EXISTS
-- Original (doesn't work):
-- SELECT name FROM users u WHERE EXISTS (SELECT 1 FROM posts p WHERE p.user_id = u.id OR u.active = 1)
-- Workaround (works):
-- Expected: 5 rows (Alice, Bob, Charlie, Eve)
SELECT name FROM users u 
WHERE u.active = 1 
   OR EXISTS (SELECT 1 FROM posts p WHERE p.user_id = u.id)
ORDER BY name;

-- Workaround 3: Use LEFT JOIN instead of EXISTS for some cases
-- Original (doesn't work):
-- SELECT name FROM users u WHERE EXISTS (SELECT 1 FROM posts p WHERE p.user_id = u.id OR u.country = 'Japan')
-- Workaround (works):
-- Expected: 4 rows (users from Japan or with posts)
SELECT DISTINCT u.name 
FROM users u 
LEFT JOIN posts p ON p.user_id = u.id
WHERE u.country = 'Japan' OR p.id IS NOT NULL
ORDER BY u.name;

-- Workaround 4: Use IN for simple cases
-- When you just need users with posts (no OR condition)
-- Expected: 3 rows (Alice, Charlie, Eve)
SELECT name FROM users
WHERE id IN (SELECT user_id FROM posts)
ORDER BY name;

-- Workaround 5: Split complex conditions
-- For NOT EXISTS with OR
-- Original (doesn't work):
-- SELECT name FROM users u WHERE NOT EXISTS (SELECT 1 FROM posts p WHERE p.user_id = u.id OR u.active = 1)
-- Workaround (works):
-- Expected: 1 row (David - not active and no posts)
SELECT name FROM users u
WHERE u.active = 0
  AND NOT EXISTS (SELECT 1 FROM posts p WHERE p.user_id = u.id)
ORDER BY name;

-- Cleanup
DROP TABLE posts;
DROP TABLE users;