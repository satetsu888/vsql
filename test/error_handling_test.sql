-- Error Handling Tests for VSQL
-- Testing various error conditions and edge cases

-- Test 1: Non-existent table operations
-- These should return errors
SELECT * FROM non_existent_table;
INSERT INTO non_existent_table (id, name) VALUES (1, 'test');
UPDATE non_existent_table SET name = 'updated' WHERE id = 1;
DELETE FROM non_existent_table WHERE id = 1;
DROP TABLE non_existent_table;

-- Test 2: Invalid column references
CREATE TABLE test_table (id int, name text, age int);
INSERT INTO test_table VALUES (1, 'Alice', 25);

-- Non-existent column in SELECT
SELECT id, name, non_existent_column FROM test_table;

-- Non-existent column in WHERE
SELECT * FROM test_table WHERE non_existent_column = 'value';

-- Non-existent column in INSERT
INSERT INTO test_table (id, name, non_existent_column) VALUES (2, 'Bob', 'invalid');

-- Non-existent column in UPDATE
UPDATE test_table SET non_existent_column = 'value' WHERE id = 1;

-- Test 3: Type mismatches and invalid comparisons
-- Comparing incompatible types
SELECT * FROM test_table WHERE name > 100;
SELECT * FROM test_table WHERE age = 'not a number';

-- Invalid operations
SELECT name + age FROM test_table;
SELECT * FROM test_table WHERE name + 10 = 35;

-- Test 4: NULL handling edge cases
INSERT INTO test_table VALUES (2, NULL, NULL);
INSERT INTO test_table VALUES (3, 'Charlie', NULL);

-- NULL comparisons (these should handle NULLs properly)
SELECT * FROM test_table WHERE name = NULL;  -- Should return no rows
SELECT * FROM test_table WHERE name IS NULL; -- Should return rows with NULL name
SELECT * FROM test_table WHERE age > NULL;   -- Should return no rows
SELECT * FROM test_table WHERE NULL = NULL;  -- Should return no rows

-- NULL in arithmetic
SELECT id, age * 2 as double_age FROM test_table; -- Should handle NULL age

-- Test 5: JOIN errors
CREATE TABLE orders (id int, user_id int, amount decimal);
INSERT INTO orders VALUES (1, 1, 100.50), (2, 1, 200.00), (3, 2, 150.75);

-- JOIN on non-existent table
SELECT * FROM test_table t
JOIN non_existent_table n ON t.id = n.user_id;

-- JOIN with non-existent columns
SELECT * FROM test_table t
JOIN orders o ON t.non_existent_column = o.user_id;

-- Ambiguous column references
SELECT id FROM test_table JOIN orders ON test_table.id = orders.user_id;

-- Test 6: Subquery errors
-- Subquery returning multiple rows where single row expected
UPDATE test_table SET age = (SELECT age FROM test_table);

-- Non-existent table in subquery
SELECT * FROM test_table WHERE id IN (SELECT id FROM non_existent_table);

-- Invalid column in subquery
SELECT * FROM test_table WHERE id IN (SELECT non_existent_column FROM orders);

-- Test 7: Aggregate function errors
-- Using non-aggregate columns without GROUP BY
SELECT name, COUNT(*) FROM test_table;

-- Invalid HAVING without GROUP BY
SELECT * FROM test_table HAVING COUNT(*) > 1;

-- Column in SELECT not in GROUP BY
SELECT id, name, COUNT(*) FROM test_table GROUP BY id;

-- Test 8: Data integrity issues
-- Duplicate primary keys (if enforced)
CREATE TABLE users_pk (id int PRIMARY KEY, name text);
INSERT INTO users_pk VALUES (1, 'User1');
INSERT INTO users_pk VALUES (1, 'User2'); -- Should fail if PK enforced

-- Foreign key violations (if enforced)
CREATE TABLE departments (id int PRIMARY KEY, name text);
CREATE TABLE employees (id int, name text, dept_id int REFERENCES departments(id));
INSERT INTO employees VALUES (1, 'John', 999); -- Non-existent department

-- Test 9: Invalid SQL syntax handled by parser
-- Missing required clauses
SELECT;
INSERT INTO test_table;
UPDATE test_table SET;
DELETE FROM;

-- Malformed expressions
SELECT * FROM test_table WHERE id =;
SELECT * FROM test_table WHERE;
SELECT FROM test_table;

-- Test 10: Edge cases with empty results
-- Division by zero
SELECT id, age / 0 as invalid_calc FROM test_table;

-- Empty IN clause
SELECT * FROM test_table WHERE id IN ();

-- Empty table operations
CREATE TABLE empty_table (id int, value text);
SELECT COUNT(*) FROM empty_table;
SELECT MAX(id) FROM empty_table;
SELECT * FROM test_table t JOIN empty_table e ON t.id = e.id;

-- Test 11: Special characters and injection attempts
-- SQL injection attempts (should be handled safely)
SELECT * FROM test_table WHERE name = 'Alice'; DROP TABLE test_table; --';
INSERT INTO test_table VALUES (4, 'Dave"; DROP TABLE test_table; --', 30);

-- Special characters in identifiers
CREATE TABLE "table-with-dashes" (id int, "column-name" text);
SELECT "column-name" FROM "table-with-dashes";

-- Test 12: Constraint violations
-- Check constraint (if supported)
CREATE TABLE age_check (id int, age int CHECK (age >= 0 AND age <= 150));
INSERT INTO age_check VALUES (1, -5);    -- Should fail
INSERT INTO age_check VALUES (2, 200);   -- Should fail

-- Not null constraint (if supported)
CREATE TABLE not_null_test (id int NOT NULL, name text NOT NULL);
INSERT INTO not_null_test VALUES (NULL, 'Name'); -- Should fail
INSERT INTO not_null_test VALUES (1, NULL);      -- Should fail

-- Test 13: Transaction errors (if transactions are supported)
BEGIN;
INSERT INTO test_table VALUES (5, 'Eve', 28);
-- Simulate error that should rollback
INSERT INTO non_existent_table VALUES (1, 'error');
COMMIT; -- Should rollback due to error

-- Test 14: Recursive queries and infinite loops
-- CTE with potential infinite recursion
WITH RECURSIVE infinite AS (
  SELECT 1 as n
  UNION ALL
  SELECT n + 1 FROM infinite WHERE n < 1000000
)
SELECT * FROM infinite; -- Should have recursion limit

-- Test 15: Memory and resource limits
-- Very large IN clause
SELECT * FROM test_table WHERE id IN (1,2,3,4,5,6,7,8,9,10); -- Add thousands more

-- Very deep nesting
SELECT * FROM test_table WHERE id IN (
  SELECT id FROM test_table WHERE id IN (
    SELECT id FROM test_table WHERE id IN (
      SELECT id FROM test_table WHERE id IN (
        SELECT id FROM test_table
      )
    )
  )
);

-- Cleanup
DROP TABLE test_table;
DROP TABLE orders;
DROP TABLE users_pk;
DROP TABLE departments;
DROP TABLE employees;
DROP TABLE empty_table;
DROP TABLE "table-with-dashes";
DROP TABLE age_check;
DROP TABLE not_null_test;