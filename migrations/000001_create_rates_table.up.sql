CREATE TABLE IF NOT EXISTS rates(
    id serial PRIMARY KEY,
    name VARCHAR(3),
    base_rate REAL,
    is_set BOOLEAN DEFAULT false,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);