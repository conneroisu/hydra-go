#!/root/.nix-profile/bin/bash
set -euo pipefail

echo "Starting Hydra Test Container..."

# Set up directories (running as root in Nix container)
echo "Setting up directories..."
mkdir -p /var/lib/postgresql/data
mkdir -p /var/lib/hydra
chmod 700 /var/lib/postgresql/data

# Initialize PostgreSQL data directory if it doesn't exist
if [ ! -d /var/lib/postgresql/data/base ]; then
    echo "Initializing PostgreSQL database..."
    # Use an existing nixbld user to run initdb (workaround for Nix limitations)
    chown -R nixbld1:nixbld /var/lib/postgresql
    RUNUSER_CMD=$(find /nix/store -name runuser -type f -path "*/bin/runuser" 2>/dev/null | head -1)
    $RUNUSER_CMD nixbld1 -c "initdb -D /var/lib/postgresql/data --auth-local=trust --auth-host=trust -U postgres"
    chown -R root:root /var/lib/postgresql  # Take back ownership to root
fi

# Start PostgreSQL in background
echo "Starting PostgreSQL..."
# Find runuser command for later use
RUNUSER_CMD=$(find /nix/store -name runuser -type f -path "*/bin/runuser" 2>/dev/null | head -1)
$RUNUSER_CMD nixbld1 -c "postgres -D /var/lib/postgresql/data -k /tmp -h localhost -p 5432" &
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
export PGUSER=postgres
createuser -h localhost -s postgres 2>/dev/null || true  # Ensure postgres superuser exists
createdb -h localhost -O postgres hydra 2>/dev/null || true

# Initialize Hydra database
export HYDRA_DBI="dbi:Pg:dbname=hydra;host=localhost;user=postgres"
mkdir -p /var/lib/hydra
if [ ! -f /var/lib/hydra/.initialized ]; then
    echo "Initializing Hydra..."
    # Set Hydra environment
    export HYDRA_DBI="dbi:Pg:dbname=hydra;host=localhost;user=postgres"
    hydra-init
    touch /var/lib/hydra/.initialized
fi

# Create admin user
echo "Creating admin user..."
export HYDRA_DBI="dbi:Pg:dbname=hydra;host=localhost;user=postgres"
hydra-create-user admin --full-name "Admin" --email admin@example.com --password admin --role admin 2>/dev/null || true

# Start health check server in background
echo "Starting health check server..."
python3 /usr/local/bin/health-check.py &
HEALTH_PID=$!

# Function to cleanup on exit
cleanup() {
    echo "Shutting down..."
    kill $HEALTH_PID 2>/dev/null || true
    kill $HYDRA_PID 2>/dev/null || true
    kill $POSTGRES_PID 2>/dev/null || true
    wait $POSTGRES_PID 2>/dev/null || true
}
trap cleanup EXIT TERM INT

# Start Hydra server
echo "Starting Hydra server on port 3000..."
echo "Admin login: admin / admin"
echo "Health check available at: http://localhost:8080/health"
export HYDRA_DBI="dbi:Pg:dbname=hydra;host=localhost;user=postgres"
hydra-server --host 0.0.0.0 --port 3000 &
HYDRA_PID=$!

# Wait for signals
wait