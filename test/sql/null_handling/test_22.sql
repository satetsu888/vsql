-- Test 22: NULL in JOIN conditions
-- Expected: NULL values in JOIN conditions should not match
-- Status: May have issues with NULL join handling

-- Setup
CREATE TABLE left_table (id INTEGER, join_key INTEGER, data TEXT);
CREATE TABLE right_table (id INTEGER, join_key INTEGER, info TEXT);

INSERT INTO left_table VALUES 
    (1, 100, 'Data 1'),
    (2, NULL, 'Data 2'),
    (3, 200, 'Data 3');

INSERT INTO right_table VALUES 
    (1, 100, 'Info A'),
    (2, NULL, 'Info B'),
    (3, 300, 'Info C');

-- Test Query: NULL = NULL in JOIN should not match
SELECT 
    l.id as left_id,
    l.join_key as left_key,
    r.id as right_id,
    r.join_key as right_key
FROM left_table l
INNER JOIN right_table r ON l.join_key = r.join_key
ORDER BY l.id;

-- Cleanup
DROP TABLE right_table;
DROP TABLE left_table;