-- Test: Boolean column in simple WHERE
-- Expected: 2 rows (Alice, Bob)

CREATE TABLE test_bool (id int, name text, active boolean);
INSERT INTO test_bool VALUES
  (1, 'Alice', true),
  (2, 'Bob', true),
  (3, 'Charlie', false);

SELECT name FROM test_bool WHERE active = true;
DROP TABLE test_bool;