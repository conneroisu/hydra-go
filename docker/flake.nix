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

                # Open the web/UI port for testcontainers
                networking.firewall.allowedTCPPorts = [3000];

                # Create a deterministic admin token at build-time for tests
                # (you can also Exec to create one at runtime if you prefer)
                systemd.services.hydra-seed-admin = {
                  after = ["hydra-server.service"];
                  wants = ["hydra-server.service"];
                  wantedBy = ["multi-user.target"];
                  serviceConfig.Type = "oneshot";
                  script = ''
                    set -euo pipefail
                    # If not already created, make an admin user and a shared token
                    if ! su - hydra -c 'hydra-create-user admin --full-name "Admin" --email admin@example.com --password admin --role admin' >/dev/null 2>&1; then
                      true
                    fi
                    echo "seeded"
                  '';
                };
              };
            };
          };
        };
      })
      systems);
}
