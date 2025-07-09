-- Test 21: NULL in string concatenation
-- Expected: String concatenation with NULL should return NULL
-- Status: May fail if string concatenation not implemented

-- Setup
CREATE TABLE test_null (id INTEGER, first_name TEXT, last_name TEXT);
INSERT INTO test_null VALUES 
    (1, 'John', 'Doe'),
    (2, 'Jane', NULL),
    (3, NULL, 'Smith');

-- Test Query
SELECT 
    id,
    first_name,
    last_name,
    first_name || ' ' || last_name as full_name
FROM test_null
ORDER BY id;

-- Cleanup
DROP TABLE test_null;