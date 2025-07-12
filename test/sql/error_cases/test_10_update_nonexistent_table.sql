-- Test: UPDATE on non-existent table returns 0 affected rows
-- Expected: no rows

-- Test Query
UPDATE non_existent_table SET name = 'Updated' WHERE id = 1;