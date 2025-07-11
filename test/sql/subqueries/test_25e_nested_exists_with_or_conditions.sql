-- Test: Nested EXISTS with OR conditions
-- Expected: Should return users who have posts with comments OR are active
-- Actual: Only returns users with posts that have comments

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

-- Test query
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

-- Cleanup
DROP TABLE comments;
DROP TABLE posts;
DROP TABLE users;