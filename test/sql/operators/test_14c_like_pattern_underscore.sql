-- Test: Pattern with _ (single character)
-- Expected: 4 rows (2,5,6,7)

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

-- Test query
SELECT id, text_val FROM pattern_test WHERE text_val LIKE 'H_llo' ORDER BY id;

-- Cleanup
DROP TABLE pattern_test;