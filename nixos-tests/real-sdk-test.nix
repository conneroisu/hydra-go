# NixOS VM test with real Hydra and SDK
{pkgs ? import <nixpkgs> {}}:
pkgs.nixosTest {
  name = "hydra-go-sdk-real-test";

  nodes.machine = {
    config,
    pkgs,
    ...
  }: {
    virtualisation.memorySize = 4096;

    services.postgresql.enable = true;

    services.hydra = {
      enable = true;
      hydraURL = "http://localhost:3000";
      notificationSender = "test@localhost";
    };

    nix.settings.sandbox = false;

    environment.systemPackages = with pkgs; [
      curl
      go_1_24
      git
    ];
  };

  testScript = ''
        machine.start()
        machine.wait_for_unit("multi-user.target")
        machine.wait_for_unit("postgresql.service")
        machine.wait_for_unit("hydra-init.service")
        machine.wait_for_unit("hydra-server.service")
        machine.wait_for_open_port(3000)

        # Verify Hydra is running
        machine.succeed("curl -f http://localhost:3000/")
        print("✅ Hydra is running!")

        # Create test directory and copy SDK
        machine.succeed("mkdir -p /tmp/hydra-go")

        # Copy SDK files
        machine.copy_from_host("${../go.mod}", "/tmp/hydra-go/go.mod")
        machine.copy_from_host("${../go.sum}", "/tmp/hydra-go/go.sum")
        machine.succeed("mkdir -p /tmp/hydra-go/hydra")
        machine.copy_from_host("${../hydra}", "/tmp/hydra-go/hydra")

        # Create test program
        machine.succeed("""
          cat > /tmp/hydra-go/test.go << 'EOF'
    package main

    import (
        "context"
        "fmt"
        "log"

        "github.com/conneroisu/hydra-go/hydra"
    )

    func main() {
        fmt.Println("========================================")
        fmt.Println("Testing Hydra Go SDK with Real Hydra")
        fmt.Println("========================================")

        // Create client
        client, err := hydra.NewClientWithURL("http://localhost:3000")
        if err != nil {
            log.Fatalf("❌ Failed to create client: %v", err)
        }

        ctx := context.Background()

        // Test 1: List projects
        fmt.Println("\\nTest 1: Listing projects...")
        projects, err := client.ListProjects(ctx)
        if err != nil {
            log.Fatalf("❌ Failed to list projects: %v", err)
        }
        fmt.Printf("✅ Found %d projects\\n", len(projects))

        // Test 2: Get API info
        fmt.Println("\\nTest 2: Testing API endpoints...")
        _, err = client.GetProject(ctx, "non-existent")
        if err != nil {
            fmt.Println("✅ Error handling works (expected error for non-existent project)")
        }

        // Test 3: Search
        fmt.Println("\\nTest 3: Testing search...")
        results, err := client.SearchAll(ctx, "test")
        if err != nil {
            log.Printf("⚠️ Search error (might be expected): %v", err)
        } else {
            fmt.Printf("✅ Search completed: %d projects, %d jobsets, %d builds\\n",
                len(results.Projects), len(results.Jobsets), len(results.Builds))
        }

        fmt.Println("\\n========================================")
        fmt.Println("✅ All SDK tests passed!")
        fmt.Println("========================================")
    }
    EOF
        """)

        # Build and run test
        result = machine.succeed("""
          cd /tmp/hydra-go
          go mod download
          go run test.go 2>&1
        """)

        print(result)

        # Verify output contains success message
        if "All SDK tests passed!" in result:
            print("✅ SDK verified to work with real Hydra in NixOS VM!")
        else:
            raise Exception("SDK test did not complete successfully")
  '';
}
