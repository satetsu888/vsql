-- Test: OR with function on outer table column
-- Expected: Should return users with posts OR users with name starting with 'B'
-- Expected: 4 rows

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
SELECT name FROM users u
WHERE EXISTS (
  SELECT 1 FROM posts p 
  WHERE p.user_id = u.id 
  OR u.name LIKE 'B%'
)
ORDER BY name;

-- Cleanup
DROP TABLE posts;
DROP TABLE users;