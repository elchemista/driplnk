-- Dev helper: create the local driplink database if it does not exist.
-- Run this against the default "postgres" database before other migrations.
CREATE EXTENSION IF NOT EXISTS dblink;

DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_database WHERE datname = 'driplink') THEN
        PERFORM dblink_exec('dbname=postgres', 'CREATE DATABASE driplink');
    END IF;
END
$$;
