-- Test: Comment syntax in string literals
-- Description: Ensure comment syntax in strings is not treated as comments
-- Expected: 1 row

CREATE TABLE comment_test (id INT, name TEXT);
INSERT INTO comment_test VALUES (1, '-- This is not a comment');
INSERT INTO comment_test VALUES (2, '/* This is also not a comment */');
SELECT * FROM comment_test WHERE name LIKE '--%' OR name LIKE '/*%';
-- Expected: 2 rows