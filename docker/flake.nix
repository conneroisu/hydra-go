{
  description = "Hydra-in-Docker image for testing";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

  outputs = {
    self,
    nixpkgs,
    ...
  }: let
    systems = ["x86_64-linux" "aarch64-linux"];
  in
    builtins.listToAttrs (map (system: let
        pkgs = import nixpkgs {inherit system;};
        
        # Create a startup script
        startupScript = pkgs.writeScript "hydra-startup.sh" ''
          #!/bin/bash
          set -euo pipefail
          
          # Initialize PostgreSQL data directory if it doesn't exist
          if [ ! -d /var/lib/postgresql/data/base ]; then
            mkdir -p /var/lib/postgresql/data
            chown postgres:postgres /var/lib/postgresql/data
            chmod 700 /var/lib/postgresql/data
            su - postgres -c "${pkgs.postgresql}/bin/initdb -D /var/lib/postgresql/data"
          fi
          
          # Start PostgreSQL
          su - postgres -c "${pkgs.postgresql}/bin/postgres -D /var/lib/postgresql/data -k /tmp" &
          POSTGRES_PID=$!
          
          # Wait for PostgreSQL to start
          for i in {1..30}; do
            if su - postgres -c "${pkgs.postgresql}/bin/pg_isready -h localhost" >/dev/null 2>&1; then
              echo "PostgreSQL is ready"
              break
            fi
            sleep 1
          done
          
          # Create Hydra database and user
          su - postgres -c "${pkgs.postgresql}/bin/createuser -s hydra" 2>/dev/null || true
          su - postgres -c "${pkgs.postgresql}/bin/createdb -O hydra hydra" 2>/dev/null || true
          
          # Initialize Hydra database
          export HYDRA_DBI="dbi:Pg:dbname=hydra;host=localhost;user=hydra"
          mkdir -p /var/lib/hydra
          if [ ! -f /var/lib/hydra/.initialized ]; then
            ${pkgs.hydra}/bin/hydra-init
            touch /var/lib/hydra/.initialized
          fi
          
          # Create admin user
          ${pkgs.hydra}/bin/hydra-create-user admin --full-name "Admin" --email admin@example.com --password admin --role admin 2>/dev/null || true
          
          # Start health check server in background
          ${pkgs.python3}/bin/python3 -c '
import http.server
import socketserver
import json
import threading
from urllib.request import urlopen

class HealthHandler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == "/health":
            try:
                response = urlopen("http://localhost:3000/", timeout=5)
                if response.status == 200:
                    self.send_response(200)
                    self.send_header("Content-type", "application/json")
                    self.end_headers()
                    self.wfile.write(json.dumps({"status": "healthy", "hydra": "running"}).encode())
                else:
                    raise Exception("Hydra not responding")
            except Exception as e:
                self.send_response(503)
                self.send_header("Content-type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps({"status": "unhealthy", "error": str(e)}).encode())
        else:
            self.send_response(404)
            self.end_headers()

def run_health_server():
    with socketserver.TCPServer(("", 8080), HealthHandler) as httpd:
        httpd.serve_forever()

threading.Thread(target=run_health_server, daemon=True).start()
' &
          
          # Start Hydra server
          exec ${pkgs.hydra}/bin/hydra-server --host 0.0.0.0 --port 3000
        '';
        
        # Health check script
        healthCheckScript = pkgs.writeScript "health-check.py" ''
          #!${pkgs.python3}/bin/python3
          import http.server
          import socketserver
          import json
          from urllib.request import urlopen

          class HealthHandler(http.server.BaseHTTPRequestHandler):
              def do_GET(self):
                  if self.path == "/health":
                      try:
                          response = urlopen("http://localhost:3000/", timeout=5)
                          if response.status == 200:
                              self.send_response(200)
                              self.send_header("Content-type", "application/json")
                              self.end_headers()
                              self.wfile.write(json.dumps({"status": "healthy", "hydra": "running"}).encode())
                          else:
                              raise Exception("Hydra not responding")
                      except Exception as e:
                          self.send_response(503)
                          self.send_header("Content-type", "application/json")
                          self.end_headers()
                          self.wfile.write(json.dumps({"status": "unhealthy", "error": str(e)}).encode())
                  else:
                      self.send_response(404)
                      self.end_headers()

          with socketserver.TCPServer(("", 8080), HealthHandler) as httpd:
              httpd.serve_forever()
        '';
      in {
        name = system;
        value = {
          packages = {
            hydraDockerImage = pkgs.dockerTools.buildImage {
              name = "ghcr.io/conneroisu/hydra-go/hydra-test";
              tag = "latest";
              
              copyToRoot = pkgs.buildEnv {
                name = "image-root";
                paths = with pkgs; [ 
                  hydra
                  postgresql
                  python3
                  curl
                  bash
                  coreutils
                  glibc
                  startupScript
                  healthCheckScript
                ];
                pathsToLink = [ "/bin" ];
              };
              
              runAsRoot = ''
                #!${pkgs.runtimeShell}
                ${pkgs.dockerTools.shadowSetup}
                
                # Create necessary directories
                mkdir -p /var/lib/postgresql
                mkdir -p /var/lib/hydra
                mkdir -p /tmp
                mkdir -p /etc/ssl/certs
                
                # Create postgres user
                groupadd -r postgres --gid=999 || true
                useradd -r -g postgres --uid=999 --home-dir=/var/lib/postgresql --shell=/bin/bash postgres || true
                chown postgres:postgres /var/lib/postgresql
                
                # Create hydra user  
                groupadd -r hydra --gid=998 || true
                useradd -r -g hydra --uid=998 --home-dir=/var/lib/hydra --shell=/bin/bash hydra || true
                chown hydra:hydra /var/lib/hydra
                
                # Set up SSL certificates
                ${pkgs.cacert}/bin/update-ca-certificates
              '';
              
              config = {
                Cmd = [ "${startupScript}" ];
                ExposedPorts = {
                  "3000/tcp" = {};
                  "8080/tcp" = {};
                };
                Env = [
                  "PATH=/bin"
                  "HYDRA_DBI=dbi:Pg:dbname=hydra;host=localhost;user=hydra"
                  "PGDATA=/var/lib/postgresql/data"
                ];
                WorkingDir = "/";
                User = "root";
              };
            };
          };
        };
      })
      systems);
}
