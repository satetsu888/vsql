-- Test: Basic correlated NOT EXISTS ✓
-- Expected: 1 row (Charlie)

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
WHERE NOT EXISTS (SELECT 1 FROM posts p WHERE p.user_id = u.id);

-- Cleanup
DROP TABLE posts;
DROP TABLE users;