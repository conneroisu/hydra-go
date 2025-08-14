#!/bin/bash
set -euo pipefail

echo "Starting Simple Hydra Test Container..."

# Initialize PostgreSQL
echo "Initializing PostgreSQL..."
su - postgres -c "initdb -D /var/lib/postgresql/data --auth-local=trust --auth-host=trust"

# Start PostgreSQL
echo "Starting PostgreSQL..."
su - postgres -c "postgres -D /var/lib/postgresql/data -k /tmp" &
POSTGRES_PID=$!

# Wait for PostgreSQL
sleep 5
until su - postgres -c "pg_isready"; do
    echo "Waiting for PostgreSQL..."
    sleep 1
done

# Create database
su - postgres -c "createdb hydra"

# Start health check server
python3 /usr/local/bin/health-check.py &
HEALTH_PID=$!

# Mock Hydra server (simple HTTP server on port 3000)
cat > /tmp/mock_hydra.py << 'EOF'
#!/usr/bin/env python3
import http.server
import socketserver
import json
import threading
import time

class MockHydraHandler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == "/":
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps({"status": "ok", "version": "mock"}).encode())
        elif self.path.startswith("/api/v1/projects"):
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            projects = [
                {"name": "nixpkgs", "displayname": "Nixpkgs", "enabled": True},
                {"name": "hydra", "displayname": "Hydra", "enabled": True}
            ]
            self.wfile.write(json.dumps(projects).encode())
        else:
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps({"status": "ok"}).encode())

    def log_message(self, format, *args):
        pass  # Suppress logs

def start_server():
    with socketserver.TCPServer(("", 3000), MockHydraHandler) as httpd:
        httpd.serve_forever()

if __name__ == "__main__":
    print("Starting mock Hydra server on port 3000...")
    start_server()
EOF

python3 /tmp/mock_hydra.py &
HYDRA_PID=$!

# Cleanup function
cleanup() {
    echo "Shutting down..."
    kill $HEALTH_PID 2>/dev/null || true
    kill $HYDRA_PID 2>/dev/null || true
    kill $POSTGRES_PID 2>/dev/null || true
    wait
}
trap cleanup EXIT TERM INT

echo "Mock Hydra server started successfully"
echo "Health check available at: http://localhost:8080/health"
echo "Hydra API mock available at: http://localhost:3000"

# Wait for signals
wait