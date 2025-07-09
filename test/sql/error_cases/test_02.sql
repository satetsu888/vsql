-- Test 2: Non-existent column in SELECT returns NULL
-- Expected: 3 rows

-- Setup
CREATE TABLE test_table (id int, name text, age int);
INSERT INTO test_table VALUES (1, 'Alice', 25);
INSERT INTO test_table VALUES (2, NULL, NULL);
INSERT INTO test_table VALUES (3, 'Charlie', NULL);

-- Test Query
SELECT id, name, non_existent_column FROM test_table;

-- Cleanup
DROP TABLE test_table;