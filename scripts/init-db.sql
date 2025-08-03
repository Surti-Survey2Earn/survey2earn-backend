-- Ensure database exists
SELECT 'CREATE DATABASE surti_db' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'surti_db')\gexec

-- Connect to the database
\c surti_db;

-- Add any additional initialization here
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create any custom types if needed
-- CREATE TYPE survey_status AS ENUM ('draft', 'active', 'completed', 'cancelled');

-- Grant permissions to maziazi user
GRANT ALL PRIVILEGES ON DATABASE surti_db TO maziazi;
GRANT ALL PRIVILEGES ON SCHEMA public TO maziazi;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO maziazi;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO maziazi;

-- Set default privileges for future objects
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO maziazi;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO maziazi;

-- Log successful initialization
SELECT 'Database initialization completed successfully' AS status;
