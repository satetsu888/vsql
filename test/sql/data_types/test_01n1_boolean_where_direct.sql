-- Test: Boolean column in WHERE without comparison
-- Expected: 3 rows
-- PostgreSQL: WHERE active should return only rows where active=true (ids: 1,2,6)

CREATE TABLE bool_std (id int, active boolean, verified boolean);
INSERT INTO bool_std VALUES 
  (1, true, true),
  (2, true, false),
  (3, false, true),
  (4, false, false),
  (5, NULL, true),
  (6, true, NULL);

SELECT id FROM bool_std WHERE active;

DROP TABLE bool_std;