-- Test: Type promotion from Integer to Float
-- Expected: 3 rows

CREATE TABLE prices (id text, price text);

-- First INSERT determines the type as Integer
INSERT INTO prices VALUES (1, 100);     -- price: Integer
INSERT INTO prices VALUES (2, 99.99);   -- price: Integer→Float promotion (allowed)
INSERT INTO prices VALUES (3, 200.50);  -- price: Float value OK

SELECT id, price FROM prices ORDER BY id;

DROP TABLE prices;