-- Test: Same test with integer (works)
-- Expected: 2 rows (Alice, Bob)

CREATE TABLE test_int (id int, name text, active int);
INSERT INTO test_int VALUES
  (1, 'Alice', 1),
  (2, 'Bob', 1),
  (3, 'Charlie', 0);

SELECT name FROM test_int WHERE active = 1;
DROP TABLE test_int;