-- Test: Type mismatch - Integer to String
-- Expected: error

CREATE TABLE users (id text, age text);

-- First INSERT determines age as Integer
INSERT INTO users VALUES (1, 25);      -- age: Integer
-- This should fail - cannot change from Integer to String
INSERT INTO users VALUES (2, 'twenty-six');  -- Error: Integer→String

DROP TABLE users;