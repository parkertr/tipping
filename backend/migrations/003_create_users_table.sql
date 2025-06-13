-- Create users table for user read model
CREATE TABLE IF NOT EXISTS users_view (
    id VARCHAR(255) PRIMARY KEY,
    google_id VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    picture_url TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    total_points INTEGER NOT NULL DEFAULT 0,
    correct_predictions INTEGER NOT NULL DEFAULT 0,
    total_predictions INTEGER NOT NULL DEFAULT 0,
    current_rank INTEGER NOT NULL DEFAULT 0
);

-- Create index on google_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_google_id ON users_view(google_id);

-- Create index on email for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON users_view(email);

-- Create index on is_active for filtering active users
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users_view(is_active);
