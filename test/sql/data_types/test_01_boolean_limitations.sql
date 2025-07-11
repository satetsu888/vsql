-- Test: Boolean type limitations
-- Status: Boolean values are not properly stored or compared
-- Workaround: Use integer (0/1) or text ('true'/'false') instead

-- Test 1: Boolean column storage
-- Expected: true and false values
-- Actual: Empty or NULL values
CREATE TABLE bool_test (id int, active boolean);
INSERT INTO bool_test VALUES (1, true), (2, false);
SELECT id, active FROM bool_test;
DROP TABLE bool_test;

-- Test 2: Boolean comparison
-- Expected: 1 row (id=1 with true)
-- Actual: 0 rows
CREATE TABLE bool_compare (id int, active boolean);
INSERT INTO bool_compare VALUES (1, true), (2, false);
SELECT id FROM bool_compare WHERE active = true;
DROP TABLE bool_compare;

-- Test 3: Boolean with IS NULL
-- Shows that boolean values might be stored as NULL
CREATE TABLE bool_null (id int, active boolean);
INSERT INTO bool_null VALUES (1, true), (2, false), (3, NULL);
SELECT id, active IS NULL as is_null FROM bool_null;
DROP TABLE bool_null;

-- Workaround 1: Use integer (0/1)
-- This works correctly
CREATE TABLE int_bool (id int, active int);
INSERT INTO int_bool VALUES (1, 1), (2, 0);
SELECT id FROM int_bool WHERE active = 1;
DROP TABLE int_bool;

-- Workaround 2: Use text ('true'/'false')
-- This also works correctly
CREATE TABLE text_bool (id int, active text);
INSERT INTO text_bool VALUES (1, 'true'), (2, 'false');
SELECT id FROM text_bool WHERE active = 'true';
DROP TABLE text_bool;

-- Note: This limitation affects any query using boolean columns,
-- including EXISTS subqueries with OR conditions that appeared to fail
-- in earlier tests (but actually work fine with proper data types).