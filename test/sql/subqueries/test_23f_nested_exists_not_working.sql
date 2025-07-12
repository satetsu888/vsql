-- Test: Nested EXISTS - This query attempts to use EXISTS inside another EXISTS
-- Expected: 2 rows (Alice, Bob - active users with posts)
-- Status: FAILING - Returns 0 rows instead of expected results

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

-- Test query
SELECT name FROM users u
WHERE EXISTS (
  SELECT 1 FROM posts p 
  WHERE p.user_id = u.id
  AND EXISTS (SELECT 1 FROM users u2 WHERE u2.id = p.user_id AND u2.active = true)
);

-- Cleanup
DROP TABLE posts;
DROP TABLE users;