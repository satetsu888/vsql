-- Test: Boolean handling limitation in EXISTS subqueries
-- Status: Boolean values don't work correctly in WHERE clauses
-- Workaround: Use integer (0/1) instead of boolean

-- Test 1: Boolean column in simple WHERE (doesn't work)
-- Expected: 2 rows (Alice, Bob)
-- Actual: 0 rows
CREATE TABLE test_bool (id int, name text, active boolean);
INSERT INTO test_bool VALUES
  (1, 'Alice', true),
  (2, 'Bob', true),
  (3, 'Charlie', false);

SELECT name FROM test_bool WHERE active = true;
DROP TABLE test_bool;

-- Test 2: Same test with integer (works)
-- Expected: 2 rows (Alice, Bob)
CREATE TABLE test_int (id int, name text, active int);
INSERT INTO test_int VALUES
  (1, 'Alice', 1),
  (2, 'Bob', 1),
  (3, 'Charlie', 0);

SELECT name FROM test_int WHERE active = 1;
DROP TABLE test_int;

-- Test 3: Boolean in EXISTS with OR (doesn't work)
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

-- Test 4: Same query with integer (works correctly!)
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

-- Test 5: String 'true'/'false' (works with proper comparison)
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

-- CONCLUSION: The actual issue is with boolean type handling, not OR conditions!
-- When using integers instead of booleans, the OR conditions with outer table
-- references work correctly in EXISTS subqueries.