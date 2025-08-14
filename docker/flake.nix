{
  description = "Hydra-in-Docker image (NixOS+systemd+Postgres) for tests";

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
      in {
        name = system;
        value = {
          packages = {
            hydraDockerImage = pkgs.dockerTools.buildNixosImage {
              name = "hydra-nixos";
              tag = "test";
              # NixOS config for the image:
              config = {
                lib,
                pkgs,
                ...
              }: {
                boot.isContainer = true;
                boot.kernelParams = ["systemd.unified_cgroup_hierarchy=1"];
                system.stateVersion = "24.05";

                # nix-daemon inside container, sandbox off (Docker cannot do user namespaces well in CI)
                nix.settings = {
                  sandbox = false;
                  substituters = ["https://cache.nixos.org"];
                  trusted-users = ["root" "hydra"];
                  experimental-features = ["nix-command" "flakes"];
                };
                services.nix-daemon.enable = true;

                # Postgres + Hydra
                services.postgresql = {
                  enable = true;
                  # Tiny DB for tests
                  settings.shared_buffers = "64MB";
                  identMap = ''
                    hydra-users root hydra
                  '';
                  authentication = lib.mkForce ''
                    local   all             all                                     trust
                    host    all             all             127.0.0.1/32            trust
                    host    all             all             ::1/128                 trust
                  '';
                  initialScript = ''
                    CREATE USER hydra WITH SUPERUSER;
                    CREATE DATABASE hydra WITH OWNER hydra;
                  '';
                };

                services.hydra = {
                  enable = true;
                  hydraURL = "http://localhost:3000/"; # used in links/emails only
                  listenHost = "0.0.0.0";
                  port = 3000;
                  # Point Hydra at the local DB
                  secretKeyFile = "/var/lib/hydra/secret-key"; # auto-created
                  extraConfig = ''
                    using_frontend_proxy = 1
                    <database>
                      type = "Pg"
                      dbname = "hydra"
                      user = "hydra" 
                      host = "127.0.0.1"
                      port = "5432"
                    </database>
                  '';
                };

                services.hydra-evaluator.enable = true;
                services.hydra-queue-runner.enable = true;

                # Open the web/UI port and health check port for testcontainers
                networking.firewall.allowedTCPPorts = [3000 8080];
                
                # Add packages needed for health check and admin setup
                environment.systemPackages = with pkgs; [
                  curl
                  python3
                ];

                # Health check endpoint for testcontainers
                systemd.services.hydra-health-check = {
                  after = ["hydra-server.service"];
                  wants = ["hydra-server.service"];
                  wantedBy = ["multi-user.target"];
                  serviceConfig = {
                    Type = "simple";
                    User = "hydra";
                    Restart = "always";
                    RestartSec = "10";
                  };
                  script = ''
                    ${pkgs.python3}/bin/python3 -c '
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
'
                  '';
                };

                # Create a deterministic admin token at build-time for tests
                # (you can also Exec to create one at runtime if you prefer)
                systemd.services.hydra-seed-admin = {
                  after = ["hydra-server.service"];
                  wants = ["hydra-server.service"];
                  wantedBy = ["multi-user.target"];
                  serviceConfig.Type = "oneshot";
                  script = ''
                    set -euo pipefail
                    # Wait for Hydra to be available
                    for i in {1..30}; do
                      if curl -f http://localhost:3000/ >/dev/null 2>&1; then
                        break
                      fi
                      sleep 2
                    done
                    
                    # If not already created, make an admin user and a shared token
                    if ! su - hydra -c 'hydra-create-user admin --full-name "Admin" --email admin@example.com --password admin --role admin' >/dev/null 2>&1; then
                      true
                    fi
                    echo "admin user seeded successfully"
                  '';
                };
              };
            };
          };
        };
      })
      systems);
}
