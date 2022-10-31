CREATE TABLE IF NOT EXISTS expenses(
    id serial PRIMARY KEY,
    user_id INT,
    amount REAL,
    category VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),

    CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Hash index on user_id column
-- user_id is the most relevant column when expenses are retrieved
-- Hash is great for fast search with = operator, which is the most relevant operation
-- B-Tree is a viable choice as well, although we would not compare by ids
CREATE INDEX idx_expenses_user_id ON expenses USING HASH (user_id);
