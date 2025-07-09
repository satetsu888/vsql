-- Test 12: NOT LIKE operator
-- Expected: 2 rows (Alice, Bobby)
-- Test: Names that don't start with 'Bob'

-- Setup
CREATE TABLE test_users (id INTEGER, name TEXT);
INSERT INTO test_users VALUES 
    (1, 'Alice'),
    (2, 'Bob'),
    (3, 'Bobby'),
    (4, 'Robert');

-- Test Query
SELECT id, name 
FROM test_users 
WHERE name NOT LIKE 'Bob%'
ORDER BY id;

-- Cleanup
DROP TABLE test_users;