-- Test: Same query with integer (works correctly!)
-- Expected: 2 rows (Alice, Bob)

CREATE TABLE users_int (id int, name text, active int);
CREATE TABLE posts_int (user_id int);
INSERT INTO users_int VALUES (1, 'Alice', 1), (2, 'Bob', 1), (3, 'Charlie', 0);
INSERT INTO posts_int VALUES (1);

SELECT name FROM users_int u
WHERE EXISTS (
  SELECT 1 FROM posts_int p 
  WHERE p.user_id = u.id OR u.active = 1
)
ORDER BY name;

DROP TABLE posts_int;
DROP TABLE users_int;

-- CONCLUSION: The actual issue is with boolean type handling, not OR conditions!
-- When using integers instead of booleans, the OR conditions with outer table
-- references work correctly in EXISTS subqueries.