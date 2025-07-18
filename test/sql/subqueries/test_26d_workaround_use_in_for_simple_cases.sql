-- Workaround 4: Use IN for simple cases
-- When you just need users with posts (no OR condition)
-- Expected: 3 rows (Alice, Charlie, Eve)

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

-- Test query
SELECT name FROM users
WHERE id IN (SELECT user_id FROM posts)
ORDER BY name;

-- Cleanup
DROP TABLE posts;
DROP TABLE users;