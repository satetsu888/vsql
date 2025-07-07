-- Simple test to understand NULL comparison behavior

CREATE TABLE test_null (id INTEGER, val INTEGER);
INSERT INTO test_null VALUES (1, 100), (2, NULL), (3, 200);

-- Test direct NULL comparison
SELECT id, val FROM test_null WHERE val = NULL;

-- Test NULL != NULL
SELECT id, val FROM test_null WHERE val != NULL;

-- Test self comparison with NULLs
SELECT id, val FROM test_null WHERE val = val;

-- Test self inequality with NULLs  
SELECT id, val FROM test_null WHERE val != val;

DROP TABLE test_null;