-- Workaround 3: Use LEFT JOIN instead of EXISTS for some cases
-- Original (doesn't work):
-- SELECT name FROM users u WHERE EXISTS (SELECT 1 FROM posts p WHERE p.user_id = u.id OR u.country = 'Japan')
-- Expected: 3 rows (users from Japan: Alice, Charlie OR users with posts: Alice, Charlie, Eve)

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
SELECT DISTINCT u.name 
FROM users u 
LEFT JOIN posts p ON p.user_id = u.id
WHERE u.country = 'Japan' OR p.id IS NOT NULL
ORDER BY u.name;

-- Cleanup
DROP TABLE posts;
DROP TABLE users;