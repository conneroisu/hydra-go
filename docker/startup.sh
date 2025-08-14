#!/bin/bash
set -euo pipefail

echo "Starting Hydra Test Container..."

# Create users dynamically
echo "Creating system users..."
groupadd -r postgres --gid=999 2>/dev/null || true
useradd -r -g postgres --uid=999 --home-dir=/var/lib/postgresql --shell=/bin/bash postgres 2>/dev/null || true
groupadd -r hydra --gid=998 2>/dev/null || true
useradd -r -g hydra --uid=998 --home-dir=/var/lib/hydra --shell=/bin/bash hydra 2>/dev/null || true

# Set up directories and permissions
chown -R postgres:postgres /var/lib/postgresql 2>/dev/null || true
chown -R hydra:hydra /var/lib/hydra 2>/dev/null || true

# Initialize PostgreSQL data directory if it doesn't exist
if [ ! -d /var/lib/postgresql/data/base ]; then
    echo "Initializing PostgreSQL database..."
    mkdir -p /var/lib/postgresql/data
    chown postgres:postgres /var/lib/postgresql/data
    chmod 700 /var/lib/postgresql/data
    su - postgres -c "initdb -D /var/lib/postgresql/data" || {
        echo "Failed to initialize PostgreSQL, trying as root..."
        initdb -D /var/lib/postgresql/data -U postgres
        chown -R postgres:postgres /var/lib/postgresql/data
    }
fi

# Start PostgreSQL in background
echo "Starting PostgreSQL..."
su - postgres -c "postgres -D /var/lib/postgresql/data -k /tmp" 2>/dev/null &
POSTGRES_PID=$!

# Wait for PostgreSQL to start
echo "Waiting for PostgreSQL to be ready..."
for i in {1..30}; do
    if pg_isready -h localhost >/dev/null 2>&1; then
        echo "PostgreSQL is ready"
        break
    fi
    sleep 1
done

# Create Hydra database and user
echo "Setting up Hydra database..."
createuser -h localhost -s hydra 2>/dev/null || true
createdb -h localhost -O hydra hydra 2>/dev/null || true

# Initialize Hydra database
export HYDRA_DBI="dbi:Pg:dbname=hydra;host=localhost;user=hydra"
mkdir -p /var/lib/hydra
if [ ! -f /var/lib/hydra/.initialized ]; then
    echo "Initializing Hydra..."
    hydra-init
    touch /var/lib/hydra/.initialized
fi

# Create admin user
echo "Creating admin user..."
hydra-create-user admin --full-name "Admin" --email admin@example.com --password admin --role admin 2>/dev/null || true

# Start health check server in background
echo "Starting health check server..."
python3 /usr/local/bin/health-check.py &
HEALTH_PID=$!

# Function to cleanup on exit
cleanup() {
    echo "Shutting down..."
    kill $HEALTH_PID 2>/dev/null || true
    kill $POSTGRES_PID 2>/dev/null || true
    wait $POSTGRES_PID 2>/dev/null || true
}
trap cleanup EXIT TERM INT

# Start Hydra server
echo "Starting Hydra server on port 3000..."
echo "Admin login: admin / admin"
echo "Health check available at: http://localhost:8080/health"
exec hydra-server --host 0.0.0.0 --port 3000