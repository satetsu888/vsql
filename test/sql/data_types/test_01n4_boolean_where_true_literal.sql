-- Test: Boolean literal true in WHERE
-- Expected: 6 rows  
-- Standard SQL: WHERE true should return all 6 rows

CREATE TABLE bool_std (id int, active boolean, verified boolean);
INSERT INTO bool_std VALUES 
  (1, true, true),
  (2, true, false),
  (3, false, true),
  (4, false, false),
  (5, NULL, true),
  (6, true, NULL);

SELECT id FROM bool_std WHERE true;

DROP TABLE bool_std;