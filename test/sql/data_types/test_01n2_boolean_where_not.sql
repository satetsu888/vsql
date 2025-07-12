-- Test: NOT operator on boolean column  
-- Expected: 2 rows
-- Standard SQL: WHERE NOT active should return rows where active=false (ids: 3,4)
-- Note: NULL values should not be included as NOT NULL = NULL (unknown)

CREATE TABLE bool_std (id int, active boolean, verified boolean);
INSERT INTO bool_std VALUES 
  (1, true, true),
  (2, true, false),
  (3, false, true),
  (4, false, false),
  (5, NULL, true),
  (6, true, NULL);

SELECT id FROM bool_std WHERE NOT active;

DROP TABLE bool_std;