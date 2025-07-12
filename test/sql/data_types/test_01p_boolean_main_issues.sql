-- Test: Main boolean implementation issues
-- Expected: 1 row  
-- Status: FAILING
-- Key issues: 
-- 1. WHERE <boolean_column> doesn't filter to only true values
-- 2. WHERE NOT <boolean_column> doesn't work
-- 3. WHERE <boolean_literal> doesn't filter correctly

CREATE TABLE bool_demo (id int, active boolean);
INSERT INTO bool_demo VALUES (1, true), (2, false);

-- This should return only id=1, but returns both rows
SELECT id FROM bool_demo WHERE active;

DROP TABLE bool_demo;