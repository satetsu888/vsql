-- Test: Boolean values in aggregate functions (COUNT)
-- Expected: 1 row

CREATE TABLE bool_agg (id int, active boolean);
INSERT INTO bool_agg VALUES (1, true), (2, false), (3, true), (4, NULL);

-- Count only true values
SELECT COUNT(*) AS total, COUNT(active) AS non_null_count 
FROM bool_agg WHERE active = true;

DROP TABLE bool_agg;