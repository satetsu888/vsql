-- Test: COALESCE function
-- Expected: 3 rows with values 100, 0, 200

CREATE TABLE test_coalesce (id int, value int);
INSERT INTO test_coalesce VALUES (1, 100), (2, NULL), (3, 200);

-- Test query - should return 100, 0, 200
SELECT id, COALESCE(value, 0) as value FROM test_coalesce ORDER BY id;

DROP TABLE test_coalesce;