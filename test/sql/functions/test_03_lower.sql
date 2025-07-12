-- Test: LOWER function
-- Expected: 3 rows with text transformed to lowercase

CREATE TABLE test_text (id int, name text);
INSERT INTO test_text VALUES (1, 'ALICE'), (2, 'Bob'), (3, 'ChArLiE');

-- Test query - should return alice, bob, charlie
SELECT id, LOWER(name) as lower_name FROM test_text ORDER BY id;

DROP TABLE test_text;