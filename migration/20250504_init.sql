CREATE TABLE questions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    text TEXT NOT NULL,
    answer TEXT NOT NULL,
    parent_id INTEGER DEFAULT NULL,
    FOREIGN KEY (parent_id) REFERENCES questions (id)
)