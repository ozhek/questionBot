CREATE TABLE IF NOT EXISTS questions (
    id SERIAL PRIMARY KEY,
    lang TEXT NOT NULL,
    text TEXT NOT NULL,
    answer TEXT NOT NULL,
    file_id TEXT NOT NULL DEFAULT '',
    file_type varchar(20) NOT NULL DEFAULT '',
    parent_id INTEGER REFERENCES questions(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS user_languages (
    user_id BIGINT PRIMARY KEY,
    lang TEXT NOT NULL
);