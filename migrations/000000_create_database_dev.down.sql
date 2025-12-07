-- Dev helper teardown: drop the driplink database if it was created.
-- Run against the default "postgres" database only in development.
CREATE EXTENSION IF NOT EXISTS dblink;

DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_database WHERE datname = 'driplink') THEN
        PERFORM dblink_exec('dbname=postgres', 'DROP DATABASE driplink');
    END IF;
END
$$;
