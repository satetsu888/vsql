-- Test: Basic type inference from INSERT values
-- Expected: 3 rows

CREATE TABLE test_types (id int, value text);

-- First INSERT determines the type
INSERT INTO test_types VALUES (1, 100);     -- value becomes integer
INSERT INTO test_types VALUES (2, 200);     -- integer value OK
INSERT INTO test_types VALUES (3, 300);     -- integer value OK

SELECT id, value FROM test_types ORDER BY id;

DROP TABLE test_types;