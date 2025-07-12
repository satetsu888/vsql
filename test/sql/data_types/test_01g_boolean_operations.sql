-- Test: Boolean operations on boolean columns
-- Expected: 1 row

CREATE TABLE bool_ops (id int, flag1 boolean, flag2 boolean);
INSERT INTO bool_ops VALUES 
  (1, true, true),
  (2, true, false),
  (3, false, true),
  (4, false, false),
  (5, true, NULL),
  (6, NULL, true);

-- Test AND operation
SELECT id FROM bool_ops WHERE flag1 = true AND flag2 = true;

DROP TABLE bool_ops;