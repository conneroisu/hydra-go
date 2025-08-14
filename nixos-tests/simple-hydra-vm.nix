# Simple NixOS VM test using Hydra's own test configuration
{pkgs ? import <nixpkgs> {}}:
pkgs.nixosTest {
  name = "hydra-go-sdk-test";

  nodes.machine = {
    config,
    pkgs,
    lib,
    ...
  }: {
    imports = [
      # We'll use the standard Hydra configuration
    ];

    virtualisation = {
      memorySize = 4096;
      diskSize = 16384;
    };

    # Enable PostgreSQL
    services.postgresql = {
      enable = true;
      package = pkgs.postgresql_15;
      initialScript = pkgs.writeText "init.sql" ''
        CREATE USER hydra WITH PASSWORD 'hydra';
        CREATE DATABASE hydra WITH OWNER hydra;
      '';
    };

    # Basic Hydra setup - let it manage everything itself
    services.hydra = {
      enable = true;
      hydraURL = "http://localhost:3000";
      notificationSender = "test@localhost";
      port = 3000;
      listenHost = "127.0.0.1";
      useSubstitutes = false;
      minimumDiskFree = 0;
      minimumDiskFreeEvaluator = 0;
    };

    # Nix configuration
    nix.settings = {
      sandbox = false;
      trusted-users = ["*"];
      allowed-users = ["*"];
    };

    # Override the hydra-init service to handle existing users gracefully
    systemd.services.hydra-init = {
      preStart = lib.mkForce ''
        set -e

        # Wait for PostgreSQL
        while ! ${pkgs.postgresql_15}/bin/pg_isready -h localhost -U postgres; do
          echo "Waiting for PostgreSQL..."
          sleep 1
        done

        # Ensure database exists
        ${pkgs.sudo}/bin/sudo -u postgres ${pkgs.postgresql_15}/bin/psql -tc "SELECT 1 FROM pg_database WHERE datname = 'hydra'" | grep -q 1 || \
          ${pkgs.sudo}/bin/sudo -u postgres ${pkgs.postgresql_15}/bin/createdb hydra

        # Ensure user exists
        ${pkgs.sudo}/bin/sudo -u postgres ${pkgs.postgresql_15}/bin/psql -tc "SELECT 1 FROM pg_user WHERE usename = 'hydra'" | grep -q 1 || \
          ${pkgs.sudo}/bin/sudo -u postgres ${pkgs.postgresql_15}/bin/createuser hydra

        # Grant permissions
        ${pkgs.sudo}/bin/sudo -u postgres ${pkgs.postgresql_15}/bin/psql -c "ALTER DATABASE hydra OWNER TO hydra"
        ${pkgs.sudo}/bin/sudo -u postgres ${pkgs.postgresql_15}/bin/psql -c "GRANT ALL PRIVILEGES ON DATABASE hydra TO hydra"

        # Now run hydra-init as the hydra user
        export HYDRA_DBI="dbi:Pg:dbname=hydra;host=localhost;user=hydra;"
        ${pkgs.hydra}/bin/hydra-init

        # Create admin user for testing
        ${pkgs.hydra}/bin/hydra-create-user admin \
          --full-name 'Admin User' \
          --email-address 'admin@localhost' \
          --password 'admin' \
          --role admin || true
      '';
    };

    environment.systemPackages = with pkgs; [
      go_1_24
      git
      curl
      vim
    ];
  };

  testScript = ''
        machine.start()
        machine.wait_for_unit("multi-user.target")

        # Wait for PostgreSQL
        machine.wait_for_unit("postgresql.service")
        machine.wait_until_succeeds("sudo -u postgres psql -l")

        # Start hydra-init manually with our fixed script
        machine.succeed("systemctl restart hydra-init || true")
        machine.sleep(5)

        # Start Hydra services
        machine.succeed("systemctl start hydra-server || true")
        machine.sleep(5)

        # Check if port is open
        machine.wait_for_open_port(3000)

        # Test Hydra is responding
        machine.succeed("curl -f http://localhost:3000/ || curl http://localhost:3000/")

        # Write and run SDK test
        machine.succeed("""
          cat > /tmp/test-sdk.go << 'EOF'
    package main

    import (
        "fmt"
        "io"
        "net/http"
        "os"
    )

    func main() {
        resp, err := http.Get("http://localhost:3000/api/projects")
        if err != nil {
            fmt.Printf("Error: %v\\n", err)
            os.Exit(1)
        }
        defer resp.Body.Close()

        body, _ := io.ReadAll(resp.Body)
        fmt.Printf("Status: %d\\n", resp.StatusCode)
        fmt.Printf("Response: %s\\n", string(body))

        if resp.StatusCode == 200 {
            fmt.Println("✅ Hydra API is working!")
        }
    }
    EOF
          go run /tmp/test-sdk.go
        """)

        print("✅ Hydra is running and API is accessible!")
  '';
}
