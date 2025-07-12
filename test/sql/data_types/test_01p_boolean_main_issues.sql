-- Test: Boolean column in WHERE clause filtering
-- Expected: 1 row  

CREATE TABLE bool_demo (id int, active boolean);
INSERT INTO bool_demo VALUES (1, true), (2, false);

-- This should return only id=1
SELECT id FROM bool_demo WHERE active;

DROP TABLE bool_demo;