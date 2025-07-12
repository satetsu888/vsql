-- Test: JOIN with non-existent table returns empty result
-- Expected: 0 rows

-- Setup
CREATE TABLE existing_table (id int, name text);
INSERT INTO existing_table VALUES (1, 'Alice');
INSERT INTO existing_table VALUES (2, 'Bob');

-- Test Query
SELECT e.*, n.* FROM existing_table e JOIN non_existent_table n ON e.id = n.id;

-- Cleanup
DROP TABLE existing_table;