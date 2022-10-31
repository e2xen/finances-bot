CREATE TABLE IF NOT EXISTS users(
    id bigint PRIMARY KEY,
    preferred_currency VARCHAR(3) NULL,
    month_limit real NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);