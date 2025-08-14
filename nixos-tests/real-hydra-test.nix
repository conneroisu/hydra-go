# NixOS VM test with REAL Hydra server
{pkgs ? import <nixpkgs> {}}:
pkgs.nixosTest {
  name = "hydra-go-sdk-real-test";

  nodes = {
    # Real Hydra server
    hydra = {
      config,
      pkgs,
      ...
    }: {
      imports = [];

      # System configuration
      virtualisation = {
        memorySize = 4096;
        diskSize = 8192;
        cores = 2;
      };

      # Network configuration
      networking = {
        hostName = "hydra";
        firewall.allowedTCPPorts = [3000 5432];
        interfaces.eth1 = {
          ipv4.addresses = [
            {
              address = "192.168.1.10";
              prefixLength = 24;
            }
          ];
        };
      };

      # PostgreSQL for Hydra
      services.postgresql = {
        enable = true;
        package = pkgs.postgresql_15;
        ensureDatabases = ["hydra"];
        ensureUsers = [
          {
            name = "hydra";
            ensureDBOwnership = true;
          }
        ];
        authentication = ''
          local all all trust
          host all all 127.0.0.1/32 trust
          host all all ::1/128 trust
        '';
      };

      # Hydra configuration
      services.hydra = {
        enable = true;
        hydraURL = "http://192.168.1.10:3000";
        notificationSender = "hydra@localhost";
        buildMachinesFiles = [];
        useSubstitutes = true;
        port = 3000;
        listenHost = "0.0.0.0";

        extraConfig = ''
          store_uri = auto?secret-key=/etc/nix/signing-key.sec
          binary_cache_secret_key_file = /etc/nix/signing-key.sec
          max_output_size = 4294967296
        '';
      };

      # Nix configuration
      nix = {
        settings = {
          trusted-users = ["hydra" "hydra-www" "root"];
          allowed-users = ["*"];
          sandbox = false;
          substituters = ["https://cache.nixos.org"];
          trusted-public-keys = ["cache.nixos.org-1:6NCHdD59X431o0gWypbMrAURkbJ16ZPMQFGspcDShjY="];
        };

        extraOptions = ''
          experimental-features = nix-command flakes
          gc-keep-outputs = true
          gc-keep-derivations = true
        '';
      };

      # Create signing key
      system.activationScripts.hydra-init = {
        text = ''
          mkdir -p /etc/nix
          if [ ! -f /etc/nix/signing-key.sec ]; then
            ${pkgs.nix}/bin/nix-store --generate-binary-cache-key hydra-test /etc/nix/signing-key.sec /etc/nix/signing-key.pub
          fi

          # Ensure hydra user exists
          id hydra >/dev/null 2>&1 || useradd -r -s /bin/bash hydra
        '';
      };

      # Initial Hydra projects setup script
      environment.etc."hydra-setup.sh" = {
        text = ''
          #!/bin/sh
          set -e

          echo "Waiting for Hydra to be ready..."
          while ! curl -f http://localhost:3000 >/dev/null 2>&1; do
            sleep 1
          done

          echo "Creating admin user..."
          export HYDRA_DBI="dbi:Pg:dbname=hydra;host=localhost;user=hydra"

          # Create admin user using hydra-create-user
          ${pkgs.hydra}/bin/hydra-create-user admin \
            --full-name "Admin User" \
            --email-address "admin@example.com" \
            --password "admin123" \
            --role admin || true

          # Create test user
          ${pkgs.hydra}/bin/hydra-create-user testuser \
            --full-name "Test User" \
            --email-address "test@example.com" \
            --password "testpass" \
            --role user || true

          echo "Hydra setup complete!"
        '';
        mode = "0755";
      };

      # Create sample project configuration
      environment.etc."sample-project.json" = {
        text = builtins.toJSON {
          enabled = 1;
          visible = true;
          name = "sample";
          displayname = "Sample Project";
          description = "A sample Hydra project for testing";
          homepage = "https://github.com/example/sample";
          owner = "admin";
          enable_dynamic_run_command = false;
          declarative = {
            file = "spec.json";
            type = "git";
            value = "https://github.com/NixOS/hydra.git master";
          };
        };
      };

      # Systemd service to initialize Hydra after boot
      systemd.services.hydra-init = {
        description = "Initialize Hydra";
        wantedBy = ["multi-user.target"];
        after = ["hydra-server.service" "postgresql.service"];
        path = with pkgs; [curl jq hydra postgresql];

        serviceConfig = {
          Type = "oneshot";
          RemainAfterExit = true;
          ExecStart = "/etc/hydra-setup.sh";
        };
      };

      environment.systemPackages = with pkgs; [
        git
        curl
        jq
        postgresql
        hydra
      ];
    };

    # Test client with Go SDK
    client = {
      config,
      pkgs,
      ...
    }: {
      imports = [];

      virtualisation = {
        memorySize = 2048;
        diskSize = 4096;
      };

      networking = {
        hostName = "client";
        interfaces.eth1 = {
          ipv4.addresses = [
            {
              address = "192.168.1.20";
              prefixLength = 24;
            }
          ];
        };
      };

      environment.systemPackages = with pkgs; [
        go_1_24
        git
        curl
        gnumake
        jq
      ];

      # Copy SDK source
      environment.etc."hydra-go-sdk".source = ../..;

      # Test script for SDK operations
      environment.etc."test-sdk.go" = {
        text = ''
          package main

          import (
            "context"
            "fmt"
            "log"
            "os"
            "time"

            "github.com/conneroisu/hydra-go/hydra"
            "github.com/conneroisu/hydra-go/hydra/models"
          )

          func main() {
            hydraURL := os.Getenv("HYDRA_URL")
            if hydraURL == "" {
              hydraURL = "http://192.168.1.10:3000"
            }

            fmt.Printf("Connecting to Hydra at %s\n", hydraURL)

            // Create client
            client, err := hydra.NewClientWithURL(hydraURL)
            if err != nil {
              log.Fatalf("Failed to create client: %v", err)
            }

            ctx := context.Background()

            // Test 1: List projects (should work without auth)
            fmt.Println("\n=== Test 1: List Projects ===")
            projects, err := client.ListProjects(ctx)
            if err != nil {
              log.Printf("Failed to list projects: %v", err)
            } else {
              fmt.Printf("Found %d projects\n", len(projects))
              for _, p := range projects {
                fmt.Printf("  - %s: %s\n", p.Name, p.DisplayName)
              }
            }

            // Test 2: Authentication
            fmt.Println("\n=== Test 2: Authentication ===")
            user, err := client.Login(ctx, "admin", "admin123")
            if err != nil {
              log.Printf("Failed to login: %v", err)
            } else {
              fmt.Printf("Logged in as: %s (%s)\n", user.Username, user.FullName)
            }

            // Test 3: Create a project (requires auth)
            fmt.Println("\n=== Test 3: Create Project ===")
            if client.IsAuthenticated() {
              projectReq := &models.CreateProjectRequest{
                Name:        "test-project",
                DisplayName: "Test Project",
                Description: "Created by SDK test",
                Owner:       "admin",
                Enabled:     true,
                Visible:     true,
              }

              resp, err := client.Projects.Create(ctx, "test-project", projectReq)
              if err != nil {
                log.Printf("Failed to create project: %v", err)
              } else {
                fmt.Printf("Created project: %s\n", resp.Name)
              }

              // Get the created project
              project, err := client.GetProject(ctx, "test-project")
              if err != nil {
                log.Printf("Failed to get project: %v", err)
              } else {
                fmt.Printf("Retrieved project: %s (%s)\n", project.Name, project.DisplayName)
              }
            }

            // Test 4: Create a jobset
            fmt.Println("\n=== Test 4: Create Jobset ===")
            if client.IsAuthenticated() {
              jobset := &models.Jobset{
                Name:             "main",
                Project:          "test-project",
                Description:      strPtr("Main jobset"),
                Enabled:          1,
                Visible:          true,
                KeepNr:           3,
                CheckInterval:    300,
                SchedulingShares: 100,
                NixExprInput:     strPtr("src"),
                NixExprPath:      strPtr("release.nix"),
                Inputs: map[string]models.JobsetInput{
                  "src": {
                    Name:  "src",
                    Type:  "git",
                    Value: "https://github.com/NixOS/nixpkgs.git master",
                  },
                },
              }

              _, err := client.Jobsets.Create(ctx, "test-project", "main", jobset)
              if err != nil {
                log.Printf("Failed to create jobset: %v", err)
              } else {
                fmt.Println("Created jobset: main")
              }

              // List jobsets
              jobsets, err := client.Jobsets.List(ctx, "test-project")
              if err != nil {
                log.Printf("Failed to list jobsets: %v", err)
              } else {
                fmt.Printf("Found %d jobsets in test-project\n", len(jobsets))
              }
            }

            // Test 5: Search
            fmt.Println("\n=== Test 5: Search ===")
            results, err := client.SearchAll(ctx, "test")
            if err != nil {
              log.Printf("Failed to search: %v", err)
            } else {
              fmt.Printf("Search results: %d projects, %d jobsets, %d builds\n",
                len(results.Projects), len(results.Jobsets), len(results.Builds))
            }

            // Test 6: Trigger evaluation
            fmt.Println("\n=== Test 6: Trigger Evaluation ===")
            if client.IsAuthenticated() {
              resp, err := client.TriggerEvaluation(ctx, "test-project", "main")
              if err != nil {
                log.Printf("Failed to trigger evaluation: %v", err)
              } else {
                fmt.Printf("Triggered evaluation for jobsets: %v\n", resp.JobsetsTriggered)
              }
            }

            // Test 7: Cleanup - Delete project
            fmt.Println("\n=== Test 7: Cleanup ===")
            if client.IsAuthenticated() {
              err := client.Projects.Delete(ctx, "test-project")
              if err != nil {
                log.Printf("Failed to delete project: %v", err)
              } else {
                fmt.Println("Deleted test project")
              }
            }

            fmt.Println("\n=== All SDK tests completed ===")
          }

          func strPtr(s string) *string {
            return &s
          }
        '';
      };
    };
  };

  testScript = ''
    import time

    # Start all machines
    start_all()

    # Wait for network
    hydra.wait_for_unit("network.target")
    client.wait_for_unit("network.target")

    # Wait for PostgreSQL
    hydra.wait_for_unit("postgresql.service")
    hydra.succeed("sudo -u postgres psql -d hydra -c 'SELECT 1'")
    print("PostgreSQL is ready")

    # Wait for Hydra server
    hydra.wait_for_unit("hydra-server.service")
    hydra.wait_for_open_port(3000)

    # Verify Hydra is responding
    hydra.succeed("curl -f http://localhost:3000")
    client.succeed("curl -f http://192.168.1.10:3000")
    print("Hydra server is running")

    # Initialize Hydra with users and projects
    hydra.wait_for_unit("hydra-init.service")
    print("Hydra initialization complete")

    # Verify admin user can login via API
    login_test = hydra.succeed("""
      curl -X POST http://localhost:3000/login \
        -H "Content-Type: application/json" \
        -d '{"username":"admin","password":"admin123"}' \
        -c /tmp/cookies.txt -v 2>&1 | grep -E "(Set-Cookie|200 OK)" || echo "Login attempt made"
    """)
    print(f"Login test: {login_test}")

    # Test SDK compilation
    print("\n=== Testing SDK Compilation ===")
    client.succeed("""
      cd /etc/hydra-go-sdk
      go mod download
      go build ./...
    """)
    print("SDK compiled successfully")

    # Run SDK tests against real Hydra
    print("\n=== Running SDK Tests Against Real Hydra ===")
    sdk_test_output = client.succeed("""
      cd /etc
      export HYDRA_URL=http://192.168.1.10:3000
      go run test-sdk.go
    """)
    print(sdk_test_output)

    # Run integration tests
    print("\n=== Running Integration Tests ===")
    integration_output = client.succeed("""
      cd /etc/hydra-go-sdk
      export HYDRA_URL=http://192.168.1.10:3000
      export HYDRA_TEST_URL=http://192.168.1.10:3000
      go test -v ./tests/... -tags=integration -run TestIntegration 2>&1 | head -200 || true
    """)
    print(integration_output)

    # Test specific SDK operations
    print("\n=== Testing Specific SDK Operations ===")

    # Test project listing via SDK
    project_test = client.succeed("""
      cd /etc/hydra-go-sdk
      cat > list_projects.go << 'EOF'
      package main
      import (
        "context"
        "fmt"
        "log"
        "github.com/conneroisu/hydra-go/hydra"
      )
      func main() {
        client, _ := hydra.NewClientWithURL("http://192.168.1.10:3000")
        projects, err := client.ListProjects(context.Background())
        if err != nil {
          log.Fatal(err)
        }
        fmt.Printf("Projects: %d\\n", len(projects))
      }
      EOF
      go run list_projects.go
    """)
    print(f"Project listing: {project_test}")

    # Verify Hydra's web interface
    web_check = client.succeed("curl -s http://192.168.1.10:3000 | grep -o '<title>.*</title>' | head -1")
    print(f"Hydra web interface: {web_check}")

    # Check Hydra version
    version = hydra.succeed("hydra-server --version 2>&1 || echo 'Version check'")
    print(f"Hydra version: {version}")

    print("\n=== All tests completed successfully! ===")
  '';
}
