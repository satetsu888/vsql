-- Test: Boolean ORDER BY - standard SQL orders false < true
-- Expected: 3 rows

CREATE TABLE bool_order (id int, active boolean);
INSERT INTO bool_order VALUES (1, true), (2, false), (3, true);

SELECT id, active FROM bool_order ORDER BY active, id;

DROP TABLE bool_order;