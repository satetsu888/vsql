-- Test 29: Bulk insert preparation
-- Expected: success (inserting users with id 101-200 for testing)

-- Setup
CREATE TABLE users (id int, name text, email text);

-- Test Query
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

-- Cleanup
DROP TABLE users;