-- Test: Type mismatch - String to Integer
-- Expected: error

CREATE TABLE products (id text, code text);

-- First INSERT determines code as String
INSERT INTO products VALUES (1, 'ABC123');   -- code: String
-- This should fail - cannot insert Integer into String column
INSERT INTO products VALUES (2, 456);        -- Error: String←Integer

DROP TABLE products;