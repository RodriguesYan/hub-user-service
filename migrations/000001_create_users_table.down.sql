-- Drop trigger
DROP TRIGGER IF EXISTS update_users_updated_at ON yanrodrigues.users;

-- Drop trigger function
DROP FUNCTION IF EXISTS yanrodrigues.update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS yanrodrigues.idx_users_last_login_at;
DROP INDEX IF EXISTS yanrodrigues.idx_users_created_at;
DROP INDEX IF EXISTS yanrodrigues.idx_users_is_active;
DROP INDEX IF EXISTS yanrodrigues.idx_users_email;

-- Drop users table
DROP TABLE IF EXISTS yanrodrigues.users;

-- Note: We don't drop the schema or UUID extension as they might be used by other services
-- DROP SCHEMA IF EXISTS yanrodrigues CASCADE;
-- DROP EXTENSION IF EXISTS "uuid-ossp";
