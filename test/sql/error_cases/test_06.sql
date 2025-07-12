-- Test 6: Empty table creation
-- Expected: 0 rows

-- Setup
CREATE TABLE empty_table (id int, value text);

-- Cleanup
DROP TABLE empty_table;