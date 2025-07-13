-- Test: Qualified column names should be preserved in SELECT output
-- Expected: 1 rows

CREATE TABLE t1 (id int, name text);
CREATE TABLE t2 (id int, value text);

INSERT INTO t1 VALUES (1, 'Alice'), (2, 'Bob');
INSERT INTO t2 VALUES (1, 'Value1'), (3, 'Value3');

-- Select with qualified column names should preserve the qualification
SELECT t1.id, t1.name, t2.id, t2.value 
FROM t1 
INNER JOIN t2 ON t1.id = t2.id;

DROP TABLE t1;
DROP TABLE t2;