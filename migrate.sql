DROP TABLE IF EXISTS my_table;

CREATE TABLE my_table (
    id INT PRIMARY KEY,
    name VARCHAR(100),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Insert or replace data
REPLACE INTO my_table (id, name) VALUES
    (1, 'Alice'),
    (2, 'Bob'),
    (3, 'Charlie');
