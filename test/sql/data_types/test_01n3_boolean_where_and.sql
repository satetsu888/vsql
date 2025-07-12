-- Test: Boolean AND without explicit comparison
-- Expected: 1 row
-- PostgreSQL: WHERE active AND verified should return only id=1

CREATE TABLE bool_std (id int, active boolean, verified boolean);
INSERT INTO bool_std VALUES 
  (1, true, true),
  (2, true, false),
  (3, false, true),
  (4, false, false),
  (5, NULL, true),
  (6, true, NULL);

SELECT id FROM bool_std WHERE active AND verified;

DROP TABLE bool_std;