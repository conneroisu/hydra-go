#!/bin/bash
set -euo pipefail

echo "Starting Hydra Test Container (Ubuntu)..."

# Source nix environment
. /root/.nix-profile/etc/profile.d/nix.sh

# Set up directories
echo "Setting up directories..."
mkdir -p /var/lib/postgresql/data
mkdir -p /var/lib/hydra
chown -R postgres:postgres /var/lib/postgresql
chown -R hydra:hydra /var/lib/hydra
chmod 700 /var/lib/postgresql/data

# Initialize PostgreSQL data directory if it doesn't exist
if [ ! -d /var/lib/postgresql/data/base ]; then
    echo "Initializing PostgreSQL database..."
    su - postgres -c "initdb -D /var/lib/postgresql/data --auth-local=trust --auth-host=trust"
fi

# Start PostgreSQL in background
echo "Starting PostgreSQL..."
su - postgres -c "postgres -D /var/lib/postgresql/data -k /tmp" &
POSTGRES_PID=$!

# Wait for PostgreSQL to start
echo "Waiting for PostgreSQL to be ready..."
for i in {1..30}; do
    if pg_isready -h localhost >/dev/null 2>&1; then
        echo "PostgreSQL is ready"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "PostgreSQL failed to start"
        exit 1
    fi
    sleep 1
done

# Create Hydra database and user
echo "Setting up Hydra database..."
su - postgres -c "createuser -s hydra" 2>/dev/null || true
su - postgres -c "createdb -O hydra hydra" 2>/dev/null || true

# Initialize Hydra database
export HYDRA_DBI="dbi:Pg:dbname=hydra;host=localhost;user=hydra"
mkdir -p /var/lib/hydra
if [ ! -f /var/lib/hydra/.initialized ]; then
    echo "Initializing Hydra..."
    su - hydra -c "cd /var/lib/hydra && hydra-init"
    touch /var/lib/hydra/.initialized
fi

# Create admin user
echo "Creating admin user..."
su - hydra -c "hydra-create-user admin --full-name 'Admin' --email admin@example.com --password admin --role admin" 2>/dev/null || true

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
su - hydra -c "cd /var/lib/hydra && hydra-server --host 0.0.0.0 --port 3000" &
HYDRA_PID=$!

# Wait for signals
wait