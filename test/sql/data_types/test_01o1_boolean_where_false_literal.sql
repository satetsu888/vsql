-- Test: Boolean literal false in WHERE
-- Expected: 0 rows
-- Standard SQL: WHERE false should return no rows

CREATE TABLE bool_issues (id int, active boolean);
INSERT INTO bool_issues VALUES (1, true), (2, false), (3, NULL);

SELECT id FROM bool_issues WHERE false;

DROP TABLE bool_issues;