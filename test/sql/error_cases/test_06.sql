-- Test 6: Empty table creation
-- Expected: success

-- Setup
CREATE TABLE empty_table (id int, value text);

-- Cleanup
DROP TABLE empty_table;