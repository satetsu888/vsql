-- Test 11: Complex ORDER BY with WHERE, LIKE, LIMIT
-- Expected: 5 rows (users 110-106 in descending order)
-- Based on: enhanced_integration/test_31.sql

-- Setup
CREATE TABLE users (id int, name text, email text);
INSERT INTO users (id, name, email) VALUES 
    (101, 'User101', 'user101@example.com'),
    (102, 'User102', 'user102@example.com'),
    (103, 'User103', 'user103@example.com'),
    (104, 'User104', 'user104@example.com'),
    (105, 'User105', 'user105@example.com'),
    (106, 'User106', 'user106@example.com'),
    (107, 'User107', 'user107@example.com'),
    (108, 'User108', 'user108@example.com'),
    (109, 'User109', 'user109@example.com'),
    (110, 'User110', 'user110@example.com');

-- Test Query
SELECT * FROM users WHERE id > 105 AND email LIKE '%@example.com' ORDER BY id DESC LIMIT 5;

-- Cleanup
DROP TABLE users;