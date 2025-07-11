-- Test: OR with outer table numeric comparison
-- Expected: 3 rows (users with high scores or activities)
-- Returns: Alice (activity), Charlie (high score + activity), David (high score)

-- Setup with working data types
CREATE TABLE users (id int, name text, status text, score int);
CREATE TABLE activities (user_id int, type text);

INSERT INTO users VALUES
  (1, 'Alice', 'active', 100),
  (2, 'Bob', 'active', 50),
  (3, 'Charlie', 'inactive', 200),
  (4, 'David', 'inactive', 25);

INSERT INTO activities VALUES
  (1, 'login'),
  (3, 'purchase');

-- Test query
SELECT name FROM users u
WHERE EXISTS (
  SELECT 1 FROM activities a 
  WHERE a.user_id = u.id 
  OR u.score > 75
)
ORDER BY name;

-- Cleanup
DROP TABLE activities;
DROP TABLE users;