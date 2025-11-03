#!/bin/bash
set -e

# PostgreSQL 17 binary paths
PG_BIN="/usr/lib/postgresql/17/bin"

# Ensure proper ownership of data directory
echo "Setting up data directory permissions..."
mkdir -p /postgresus-data/pgdata
chown -R postgres:postgres /postgresus-data

# Initialize PostgreSQL if not already initialized
if [ ! -s "/postgresus-data/pgdata/PG_VERSION" ]; then
    echo "Initializing PostgreSQL database..."
    gosu postgres $PG_BIN/initdb -D /postgresus-data/pgdata --encoding=UTF8 --locale=C.UTF-8
    
    # Configure PostgreSQL
    echo "host all all 127.0.0.1/32 md5" >> /postgresus-data/pgdata/pg_hba.conf
    echo "local all all trust" >> /postgresus-data/pgdata/pg_hba.conf
    echo "port = 5437" >> /postgresus-data/pgdata/postgresql.conf
    echo "listen_addresses = 'localhost'" >> /postgresus-data/pgdata/postgresql.conf
    echo "shared_buffers = 256MB" >> /postgresus-data/pgdata/postgresql.conf
    echo "max_connections = 100" >> /postgresus-data/pgdata/postgresql.conf
fi

# Start PostgreSQL in background
echo "Starting PostgreSQL..."
gosu postgres $PG_BIN/postgres -D /postgresus-data/pgdata -p 5437 &
POSTGRES_PID=$!

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
for i in {1..30}; do
    if gosu postgres $PG_BIN/pg_isready -p 5437 -h localhost >/dev/null 2>&1; then
        echo "PostgreSQL is ready!"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "PostgreSQL failed to start"
        exit 1
    fi
    sleep 1
done

# Create database and set password for postgres user
echo "Setting up database and user..."
gosu postgres $PG_BIN/psql -p 5437 -h localhost -d postgres << 'SQL'
ALTER USER postgres WITH PASSWORD 'Q1234567';
SELECT 'CREATE DATABASE postgresus OWNER postgres'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'postgresus')
\gexec
\q
SQL

# Start the main application
echo "Starting Postgresus application..."
exec ./main