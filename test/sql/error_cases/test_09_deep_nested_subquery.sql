-- Test 9: Deep nested subquery
-- Expected: 3 rows

-- Setup
CREATE TABLE test_table (id int, name text, age int);
INSERT INTO test_table VALUES (1, 'Alice', 25);
INSERT INTO test_table VALUES (2, NULL, NULL);
INSERT INTO test_table VALUES (3, 'Charlie', NULL);

-- Test Query
SELECT * FROM test_table WHERE id IN (
  SELECT id FROM test_table WHERE id IN (
    SELECT id FROM test_table
  )
);

-- Cleanup
DROP TABLE test_table;