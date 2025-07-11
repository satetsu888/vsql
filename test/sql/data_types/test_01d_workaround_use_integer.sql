-- Workaround 1: Use integer (0/1)
-- This works correctly

CREATE TABLE int_bool (id int, active int);
INSERT INTO int_bool VALUES (1, 1), (2, 0);
SELECT id FROM int_bool WHERE active = 1;
DROP TABLE int_bool;