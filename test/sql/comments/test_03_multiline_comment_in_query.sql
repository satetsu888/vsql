-- Test: Multi-line comment in query
-- Description: Test multi-line comment support
-- Expected: 1 row

CREATE TABLE comment_test (id INT, name TEXT);
INSERT INTO comment_test VALUES (1, 'Bob');
SELECT /* inline comment */ id, name FROM comment_test WHERE id = 1; /* end comment */
-- Expected: 1 row