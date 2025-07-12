-- Test: Nested EXISTS with OR referencing outer table
-- Expected: 1 row (Charlie - has purchase activity with points > 30)

-- Setup with working data types
CREATE TABLE users (id int, name text, status text, score int);
CREATE TABLE activities (user_id int, type text);
CREATE TABLE events (activity_type text, points int);

INSERT INTO users VALUES
  (1, 'Alice', 'active', 100),
  (2, 'Bob', 'active', 50),
  (3, 'Charlie', 'inactive', 200),
  (4, 'David', 'inactive', 25);

INSERT INTO activities VALUES
  (1, 'login'),
  (3, 'purchase');

INSERT INTO events VALUES ('login', 10), ('purchase', 50);

-- Test query
SELECT name FROM users u
WHERE EXISTS (
  SELECT 1 FROM activities a 
  WHERE a.user_id = u.id
  AND EXISTS (
    SELECT 1 FROM events e 
    WHERE e.activity_type = a.type 
    AND (e.points > 30 OR u.score > 150)
  )
)
ORDER BY name;

-- Cleanup
DROP TABLE events;
DROP TABLE activities;
DROP TABLE users;