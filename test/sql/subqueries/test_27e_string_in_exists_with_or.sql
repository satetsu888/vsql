-- Test: String 'true'/'false' (works with proper comparison)
-- Expected: 2 rows (Alice, Bob)

CREATE TABLE users_str (id int, name text, active text);
CREATE TABLE posts_str (user_id int);
INSERT INTO users_str VALUES (1, 'Alice', 'true'), (2, 'Bob', 'true'), (3, 'Charlie', 'false');
INSERT INTO posts_str VALUES (1);

SELECT name FROM users_str u
WHERE EXISTS (
  SELECT 1 FROM posts_str p 
  WHERE p.user_id = u.id OR u.active = 'true'
)
ORDER BY name;

DROP TABLE posts_str;
DROP TABLE users_str;