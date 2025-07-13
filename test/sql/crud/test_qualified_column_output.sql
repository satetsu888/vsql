-- Test: Qualified column names in SELECT output are preserved
-- Expected: 2 rows

CREATE TABLE users (id int, name text);
INSERT INTO users VALUES (1, 'Alice'), (2, 'Bob');

-- When using qualified column names in SELECT, they should be preserved in output
SELECT users.id, users.name FROM users ORDER BY users.id;

DROP TABLE users;