-- PostgreSQL Migration: Add user authentication system
-- Version: 001
-- Description: Initial user table and authentication setup

-- Create users table with basic authentication fields
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    salt VARCHAR(32) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    email_verified BOOLEAN DEFAULT FALSE,
    role VARCHAR(20) DEFAULT 'user' CHECK (role IN ('admin', 'moderator', 'user'))
);

-- Create user_sessions table for session management
CREATE TABLE user_sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_token VARCHAR(255) UNIQUE NOT NULL,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_active BOOLEAN DEFAULT TRUE
);

-- Create indexes for performance
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_active ON users(is_active);
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_token ON user_sessions(session_token);
CREATE INDEX idx_user_sessions_expires ON user_sessions(expires_at) WHERE is_active = TRUE;

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for automatic timestamp updates
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Insert default admin user
INSERT INTO users (username, email, password_hash, salt, role, email_verified)
VALUES (
    'admin',
    'admin@example.com',
    '$2b$10$abcdefghijklmnopqrstuvwx', -- placeholder hash
    'randomsalt1234567890123456',
    'admin',
    TRUE
);

-- Create view for active users
CREATE VIEW active_users AS
SELECT id, username, email, created_at, last_login_at, role
FROM users
WHERE is_active = TRUE;

-- Grant permissions
GRANT SELECT, INSERT, UPDATE ON users TO app_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON user_sessions TO app_user;
GRANT USAGE ON SEQUENCE users_id_seq TO app_user;
GRANT USAGE ON SEQUENCE user_sessions_id_seq TO app_user;