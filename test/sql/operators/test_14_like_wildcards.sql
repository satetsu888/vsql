-- Test 14: LIKE with different wildcard patterns
-- Expected: Tests various wildcard patterns with % and _

-- Setup
CREATE TABLE pattern_test (id INTEGER, text_val TEXT);
INSERT INTO pattern_test VALUES 
    (1, 'Hello World'),
    (2, 'Hello'),
    (3, 'World Hello'),
    (4, 'HeLLo'),
    (5, 'H_llo'),
    (6, 'Hallo'),
    (7, 'Hullo'),
    (8, 'Hi'),
    (9, NULL),
    (10, 'Hello!');

-- Test 1: Pattern with % at end
-- Expected: 8 rows (all starting with H: 1,2,4,5,6,7,8,10)
SELECT id, text_val FROM pattern_test WHERE text_val LIKE 'H%' ORDER BY id;

-- Test 2: Pattern with % at beginning and end
-- Expected: 2 rows (1,3)
SELECT id, text_val FROM pattern_test WHERE text_val LIKE '%World%' ORDER BY id;

-- Test 3: Pattern with _ (single character)
-- Expected: 3 rows (2,6,7)
SELECT id, text_val FROM pattern_test WHERE text_val LIKE 'H_llo' ORDER BY id;

-- Test 4: Pattern with both % and _
-- Expected: 6 rows (1,2,5,6,7,10 - all starting with H and containing 'l_o')
SELECT id, text_val FROM pattern_test WHERE text_val LIKE 'H%l_o%' ORDER BY id;

-- Cleanup
DROP TABLE pattern_test;