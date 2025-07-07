-- Simple Error Handling Tests for VSQL
-- Testing basic error conditions

-- Setup
CREATE TABLE test_table (id int, name text, age int);
INSERT INTO test_table VALUES (1, 'Alice', 25);
INSERT INTO test_table VALUES (2, NULL, NULL);
INSERT INTO test_table VALUES (3, 'Charlie', NULL);

CREATE TABLE orders (id int, user_id int, amount decimal);
INSERT INTO orders VALUES (1, 1, 100.50), (2, 1, 200.00), (3, 2, 150.75);

-- Test 1: Non-existent table
-- Expected: error
SELECT * FROM non_existent_table;

-- Test 2: Non-existent column in SELECT returns NULL
-- Expected: 3 rows
SELECT id, name, non_existent_column FROM test_table;

-- Test 3: Non-existent column in WHERE
-- Expected: 0 rows
SELECT * FROM test_table WHERE non_existent_column = 'value';

-- Test 4: NULL = NULL
-- Expected: 0 rows
SELECT * FROM test_table WHERE name = NULL;

-- Test 5: IS NULL
-- Expected: 1 rows
SELECT * FROM test_table WHERE name IS NULL;

-- Test 6: Empty table COUNT
CREATE TABLE empty_table (id int, value text);

-- Test 7: COUNT on empty table
-- Expected: 1 rows
SELECT COUNT(*) FROM empty_table;

-- Test 8: JOIN with empty table
-- Expected: 0 rows
SELECT * FROM test_table t JOIN empty_table e ON t.id = e.id;

-- Test 9: Deep nested subquery
-- Expected: 3 rows
SELECT * FROM test_table WHERE id IN (
  SELECT id FROM test_table WHERE id IN (
    SELECT id FROM test_table
  )
);

-- Cleanup
DROP TABLE test_table;
DROP TABLE orders;
DROP TABLE empty_table;