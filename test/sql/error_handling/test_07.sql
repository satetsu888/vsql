-- Test 7: COUNT on empty table
-- Expected: 1 rows

-- Setup
CREATE TABLE empty_table (id int, value text);

-- Test Query
SELECT COUNT(*) FROM empty_table;

-- Cleanup
DROP TABLE empty_table;