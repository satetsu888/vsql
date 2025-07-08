-- Test 3: Non-existent column in WHERE
-- Expected: 0 rows

-- Setup
CREATE TABLE test_table (id int, name text, age int);
INSERT INTO test_table VALUES (1, 'Alice', 25);
INSERT INTO test_table VALUES (2, NULL, NULL);
INSERT INTO test_table VALUES (3, 'Charlie', NULL);

-- Test Query
SELECT * FROM test_table WHERE non_existent_column = 'value';

-- Cleanup
DROP TABLE test_table;