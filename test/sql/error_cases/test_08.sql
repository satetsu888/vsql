-- Test 8: JOIN with empty table
-- Expected: 0 rows

-- Setup
CREATE TABLE test_table (id int, name text, age int);
INSERT INTO test_table VALUES (1, 'Alice', 25);
INSERT INTO test_table VALUES (2, NULL, NULL);
INSERT INTO test_table VALUES (3, 'Charlie', NULL);

CREATE TABLE empty_table (id int, value text);

-- Test Query
SELECT * FROM test_table t JOIN empty_table e ON t.id = e.id;

-- Cleanup
DROP TABLE test_table;
DROP TABLE empty_table;