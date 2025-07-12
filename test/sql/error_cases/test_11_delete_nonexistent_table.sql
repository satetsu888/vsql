-- Test: DELETE from non-existent table returns 0 affected rows
-- Expected: no rows

-- Test Query
DELETE FROM non_existent_table WHERE id = 1;