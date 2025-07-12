-- Test: ORDER BY with confirmed string type
-- Expected: 4 rows

CREATE TABLE codes (id text, code text);

-- First establish code as string type with non-numeric value
INSERT INTO codes VALUES (1, 'ABC');
-- Now insert numeric-looking strings
INSERT INTO codes VALUES (2, '100');
INSERT INTO codes VALUES (3, '20');
INSERT INTO codes VALUES (4, '3');

-- Should sort as strings: 100, 20, 3, ABC
SELECT id, code FROM codes ORDER BY code;

DROP TABLE codes;