-- Test: Multi-line comment spanning multiple lines
-- Description: Test multi-line comments that span multiple lines
-- Expected: 1 row

CREATE TABLE comment_test (id INT, name TEXT);
/*
   This is a multi-line comment
   that spans multiple lines
   and should be ignored
*/
INSERT INTO comment_test VALUES (1, 'Alice');
/* Another
   multi-line
   comment */
INSERT INTO comment_test VALUES (2, 'Bob');
SELECT COUNT(*) AS count FROM comment_test;
-- Expected: 1 row