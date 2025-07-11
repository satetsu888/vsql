-- Test: Boolean in EXISTS with OR (doesn't work)
-- Expected: 2 rows (Alice, Bob - all active users)
-- Actual: 1 row (only Alice who has posts)

CREATE TABLE users_bool (id int, name text, active boolean);
CREATE TABLE posts_bool (user_id int);
INSERT INTO users_bool VALUES (1, 'Alice', true), (2, 'Bob', true), (3, 'Charlie', false);
INSERT INTO posts_bool VALUES (1);

SELECT name FROM users_bool u
WHERE EXISTS (
  SELECT 1 FROM posts_bool p 
  WHERE p.user_id = u.id OR u.active = true
)
ORDER BY name;

DROP TABLE posts_bool;
DROP TABLE users_bool;