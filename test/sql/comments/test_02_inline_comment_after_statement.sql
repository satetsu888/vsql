-- Test: INSERT with inline comment
-- Description: Test inline comments after SQL statements
-- Expected: 1 row

CREATE TABLE comment_test (id INT, name TEXT);
INSERT INTO comment_test VALUES (1, 'Alice'); -- First user
SELECT COUNT(*) AS count FROM comment_test;
-- Expected: 1 row