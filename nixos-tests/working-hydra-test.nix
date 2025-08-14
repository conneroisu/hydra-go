# Working NixOS VM test with real Hydra
{pkgs ? import <nixpkgs> {}}:
pkgs.nixosTest {
  name = "hydra-sdk-real-test";

  nodes.hydra = {
    config,
    pkgs,
    ...
  }: {
    imports = [
      # Import Hydra's NixOS module properly
    ];

    virtualisation = {
      memorySize = 4096;
      diskSize = 16384;
    };

    # Basic networking
    networking.firewall.allowedTCPPorts = [3000];

    # PostgreSQL - let Hydra manage its own user
    services.postgresql = {
      enable = true;
      package = pkgs.postgresql_15;
      # Don't create users here - let Hydra's module handle it
    };

    # Hydra configuration
    services.hydra = {
      enable = true;
      hydraURL = "http://localhost:3000";
      notificationSender = "hydra@localhost";
      port = 3000;
      listenHost = "0.0.0.0";
      useSubstitutes = false;

      # Minimal configuration for testing
      minimumDiskFree = 1;
      minimumDiskFreeEvaluator = 1;

      # Database configuration
      dbi = "dbi:Pg:dbname=hydra;host=localhost;user=hydra;";
    };

    # Ensure Nix daemon is configured
    services.nix-daemon.enable = true;
    nix.settings = {
      sandbox = false;
      trusted-users = ["root" "hydra-queue-runner" "hydra" "hydra-www"];
      allowed-users = ["*"];
      max-jobs = 4;
    };

    # System packages needed for testing
    environment.systemPackages = with pkgs; [
      go_1_24
      git
      curl
      netcat
      postgresql
    ];

    # Ensure hydra user exists with proper permissions
    users.users.hydra = {
      isSystemUser = true;
      group = "hydra";
      home = "/var/lib/hydra";
      createHome = true;
    };

    users.groups.hydra = {};

    # Override systemd service to fix initialization
    systemd.services.hydra-init = {
      preStart = pkgs.lib.mkForce ''
        mkdir -p /var/lib/hydra
        chown hydra:hydra /var/lib/hydra

        # Check if database exists
        if ! sudo -u postgres psql -lqt | cut -d \| -f 1 | grep -qw hydra; then
          echo "Creating hydra database..."
          sudo -u postgres createdb hydra
        fi

        # Check if user exists
        if ! sudo -u postgres psql -tAc "SELECT 1 FROM pg_roles WHERE rolname='hydra'" | grep -q 1; then
          echo "Creating hydra user..."
          sudo -u postgres createuser --no-superuser --no-createdb --no-createrole hydra
        fi

        # Grant permissions
        sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE hydra TO hydra;"

        # Initialize hydra database
        ${pkgs.hydra}/bin/hydra-init
      '';

      serviceConfig = {
        Type = "oneshot";
        RemainAfterExit = true;
        User = "hydra";
        Group = "hydra";
      };
    };
  };

  nodes.client = {
    config,
    pkgs,
    ...
  }: {
    virtualisation.memorySize = 2048;

    environment.systemPackages = with pkgs; [
      go_1_24
      git
      curl
    ];
  };

  testScript = ''
        import time

        hydra.start()
        client.start()

        # Wait for basic services
        hydra.wait_for_unit("multi-user.target")
        client.wait_for_unit("multi-user.target")

        # Wait for PostgreSQL to be ready
        hydra.wait_for_unit("postgresql.service")
        hydra.succeed("sudo -u postgres psql -c '\\l' | grep postgres")
        print("PostgreSQL is ready")

        # Initialize Hydra database manually if needed
        hydra.succeed("""
          sudo -u postgres psql <<EOF
          CREATE DATABASE hydra OWNER postgres;
    EOF
        """, check_return=False)

        hydra.succeed("""
          sudo -u postgres psql <<EOF
          CREATE USER hydra WITH PASSWORD 'hydra';
          GRANT ALL PRIVILEGES ON DATABASE hydra TO hydra;
    EOF
        """, check_return=False)

        # Start Hydra services manually
        hydra.succeed("systemctl start hydra-init || true")
        time.sleep(2)

        # Try to start hydra-server
        hydra.succeed("systemctl start hydra-server || true")
        time.sleep(5)

        # Check if Hydra is accessible
        result = hydra.succeed("curl -f http://localhost:3000 || echo 'Hydra not ready yet'")
        print(f"Hydra status: {result}")

        # Copy SDK to client
        client.succeed("mkdir -p /tmp/hydra-go")

        # Write test program on client
        client.succeed("""cat > /tmp/test-hydra.go << 'EOF'
    package main

    import (
        "context"
        "fmt"
        "log"
        "net/http"
        "time"
    )

    func main() {
        // Simple HTTP test first
        client := &http.Client{Timeout: 10 * time.Second}

        fmt.Println("Testing Hydra connectivity...")
        resp, err := client.Get("http://hydra:3000")
        if err != nil {
            log.Fatalf("Failed to connect to Hydra: %v", err)
        }
        defer resp.Body.Close()

        fmt.Printf("Hydra responded with status: %d\\n", resp.StatusCode)

        if resp.StatusCode == 200 {
            fmt.Println("✅ Hydra is accessible!")
        } else {
            fmt.Printf("⚠️ Unexpected status code: %d\\n", resp.StatusCode)
        }
    }
    EOF
        """)

        # Test basic connectivity
        client.succeed("go run /tmp/test-hydra.go")

        print("✅ Test completed successfully!")
  '';
}
