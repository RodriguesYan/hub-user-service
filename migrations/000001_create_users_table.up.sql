-- Create yanrodrigues schema if it doesn't exist
CREATE SCHEMA IF NOT EXISTS yanrodrigues;

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create users table
CREATE TABLE IF NOT EXISTS yanrodrigues.users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT true NOT NULL,
    email_verified BOOLEAN DEFAULT false NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    last_login_at TIMESTAMP WITH TIME ZONE,
    locked_until TIMESTAMP WITH TIME ZONE,
    failed_login_attempts INTEGER DEFAULT 0 NOT NULL,
    
    CONSTRAINT email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    CONSTRAINT failed_attempts_non_negative CHECK (failed_login_attempts >= 0),
    CONSTRAINT first_name_not_empty CHECK (first_name <> ''),
    CONSTRAINT last_name_not_empty CHECK (last_name <> '')
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_email ON yanrodrigues.users(email);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON yanrodrigues.users(is_active);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON yanrodrigues.users(created_at);
CREATE INDEX IF NOT EXISTS idx_users_last_login_at ON yanrodrigues.users(last_login_at);

-- Create trigger to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION yanrodrigues.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON yanrodrigues.users
    FOR EACH ROW
    EXECUTE FUNCTION yanrodrigues.update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE yanrodrigues.users IS 'User accounts for the Hub Investments platform';
COMMENT ON COLUMN yanrodrigues.users.id IS 'Unique identifier for the user';
COMMENT ON COLUMN yanrodrigues.users.email IS 'User email address (unique)';
COMMENT ON COLUMN yanrodrigues.users.password IS 'Hashed user password (bcrypt)';
COMMENT ON COLUMN yanrodrigues.users.first_name IS 'User first name';
COMMENT ON COLUMN yanrodrigues.users.last_name IS 'User last name';
COMMENT ON COLUMN yanrodrigues.users.is_active IS 'Whether the user account is active';
COMMENT ON COLUMN yanrodrigues.users.email_verified IS 'Whether the user has verified their email';
COMMENT ON COLUMN yanrodrigues.users.failed_login_attempts IS 'Number of consecutive failed login attempts';
COMMENT ON COLUMN yanrodrigues.users.locked_until IS 'Timestamp until which the account is locked (null if not locked)';
