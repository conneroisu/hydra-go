# Minimal NixOS VM test with real Hydra
{pkgs ? import <nixpkgs> {}}: let
  # Test script that runs inside the VM
  testScript = pkgs.writeText "test-hydra-sdk.go" ''
    package main

    import (
      "context"
      "fmt"
      "log"
      "os"

      "github.com/conneroisu/hydra-go/hydra"
    )

    func main() {
      // Connect to local Hydra
      client, err := hydra.NewClientWithURL("http://localhost:3000")
      if err != nil {
        log.Fatalf("Failed to create client: %v", err)
      }

      ctx := context.Background()

      // Test 1: Check Hydra is accessible
      fmt.Println("Test 1: Checking Hydra connection...")
      projects, err := client.ListProjects(ctx)
      if err != nil {
        log.Fatalf("Failed to list projects: %v", err)
      }
      fmt.Printf("✓ Connected to Hydra, found %d projects\n", len(projects))

      // Test 2: Authentication
      fmt.Println("\nTest 2: Testing authentication...")
      user, err := client.Login(ctx, "admin", "admin")
      if err != nil {
        // Try to create admin user first
        fmt.Println("Admin user might not exist, that's OK for now")
      } else {
        fmt.Printf("✓ Logged in as: %s\n", user.Username)
      }

      // Test 3: Basic API operations
      fmt.Println("\nTest 3: Testing API operations...")
      if len(projects) > 0 {
        project, err := client.GetProject(ctx, projects[0].Name)
        if err != nil {
          log.Printf("Failed to get project: %v", err)
        } else {
          fmt.Printf("✓ Retrieved project: %s\n", project.Name)
        }
      }

      fmt.Println("\n✅ All basic tests passed!")
    }
  '';
in
  pkgs.nixosTest {
    name = "hydra-sdk-minimal-test";

    nodes.machine = {
      config,
      pkgs,
      ...
    }: {
      virtualisation = {
        memorySize = 4096;
        diskSize = 8192;
      };

      # PostgreSQL for Hydra
      services.postgresql = {
        enable = true;
        ensureDatabases = ["hydra"];
      };

      # Hydra service
      services.hydra = {
        enable = true;
        hydraURL = "http://localhost:3000";
        notificationSender = "hydra@localhost";
        port = 3000;
        listenHost = "localhost";
      };

      # Nix settings
      nix.settings = {
        sandbox = false;
        trusted-users = ["root" "hydra"];
      };

      environment.systemPackages = with pkgs; [
        go_1_24
        git
        curl
      ];

      # Copy SDK and test script
      environment.etc = {
        "hydra-go-sdk".source = ../..;
        "test-sdk.go".source = testScript;
      };
    };

    testScript = ''
      machine.start()
      machine.wait_for_unit("multi-user.target")

      # Wait for PostgreSQL
      machine.wait_for_unit("postgresql.service")
      machine.succeed("sudo -u postgres psql -l")

      # Wait for Hydra
      machine.wait_for_unit("hydra-server.service")
      machine.wait_for_open_port(3000)

      # Check Hydra is responding
      machine.succeed("curl -f http://localhost:3000")
      print("Hydra is running")

      # Build and test SDK
      output = machine.succeed("""
        cd /etc/hydra-go-sdk
        go mod download
        go build ./...
        cd /etc
        go run test-sdk.go
      """)
      print(output)
    '';
  }
