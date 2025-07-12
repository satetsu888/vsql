-- Test: UPPER and LOWER functions
-- Expected: 3 rows with text transformed

CREATE TABLE test_text (id int, name text);
INSERT INTO test_text VALUES (1, 'Alice'), (2, 'bob'), (3, 'ChArLiE');

-- Test query - should return ALICE, BOB, CHARLIE
SELECT id, UPPER(name) as upper_name FROM test_text ORDER BY id;

DROP TABLE test_text;