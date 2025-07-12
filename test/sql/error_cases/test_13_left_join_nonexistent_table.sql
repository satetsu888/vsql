-- Test: LEFT JOIN with non-existent table returns all rows from left table
-- Expected: 2 rows

-- Setup
CREATE TABLE existing_table (id int, name text);
INSERT INTO existing_table VALUES (1, 'Alice');
INSERT INTO existing_table VALUES (2, 'Bob');

-- Test Query
SELECT e.id, e.name FROM existing_table e LEFT JOIN non_existent_table n ON e.id = n.id;

-- Cleanup
DROP TABLE existing_table;