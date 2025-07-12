-- Test: Workaround - Use text ('true'/'false') - This also works correctly
-- Expected: 1 row

CREATE TABLE text_bool (id int, active text);
INSERT INTO text_bool VALUES (1, 'true'), (2, 'false');
SELECT id FROM text_bool WHERE active = 'true';
DROP TABLE text_bool;

-- Note: This limitation affects any query using boolean columns,
-- including EXISTS subqueries with OR conditions that appeared to fail
-- in earlier tests (but actually work fine with proper data types).