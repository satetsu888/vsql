-- Test: ORDER BY with type information
-- Expected: 3 rows

CREATE TABLE scores (id text, score text);

-- Insert numeric values that would sort incorrectly as strings
INSERT INTO scores VALUES (1, '100');
INSERT INTO scores VALUES (2, '20');
INSERT INTO scores VALUES (3, '3');

-- Should sort numerically: 3, 20, 100
SELECT id, score FROM scores ORDER BY score;

DROP TABLE scores;