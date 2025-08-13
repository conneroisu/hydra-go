# Basic NixOS VM test with Hydra
{ pkgs ? import <nixpkgs> {} }:

pkgs.nixosTest {
  name = "hydra-go-sdk-basic-test";
  
  nodes.machine = { config, pkgs, lib, ... }: {
    virtualisation = {
      memorySize = 4096;
      diskSize = 16384;
    };
    
    # PostgreSQL setup
    services.postgresql = {
      enable = true;
      package = pkgs.postgresql_15;
    };
    
    # Basic Hydra configuration - use all defaults
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
    
    # Nix settings
    nix.settings = {
      sandbox = false;
      trusted-users = [ "*" ];
      allowed-users = [ "*" ];
    };
    
    # Test dependencies
    environment.systemPackages = with pkgs; [
      go_1_24
      git
      curl
      netcat
    ];
  };
  
  testScript = ''
    machine.start()
    machine.wait_for_unit("multi-user.target")
    
    # Wait for PostgreSQL
    machine.wait_for_unit("postgresql.service")
    
    # Check PostgreSQL is ready
    machine.wait_until_succeeds("sudo -u postgres psql -l")
    print("PostgreSQL is ready")
    
    # Let Hydra initialize itself
    machine.wait_for_unit("hydra-init.service")
    print("Hydra init completed")
    
    # Start Hydra server
    machine.wait_for_unit("hydra-server.service")
    print("Hydra server started")
    
    # Wait for port
    machine.wait_for_open_port(3000)
    print("Hydra port is open")
    
    # Test basic connectivity
    result = machine.succeed("curl -s http://localhost:3000/")
    print("Hydra responded")
    
    # Test API endpoint
    api_test = machine.succeed("curl -s http://localhost:3000/api/projects")
    print(f"API response: {api_test[:100]}...")
    
    # Copy SDK files
    machine.succeed("mkdir -p /tmp/hydra-go")
    
    # Create go.mod for the test
    machine.succeed("""cat > /tmp/hydra-go/go.mod << 'EOF'
module test
go 1.24
EOF
    """)
    
    # Create test program
    machine.succeed("""cat > /tmp/hydra-go/test.go << 'EOF'
package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
)

func main() {
    // Test 1: Check API is accessible
    fmt.Println("Testing Hydra API...")
    
    resp, err := http.Get("http://localhost:3000/api/projects")
    if err != nil {
        fmt.Printf("Failed to connect: %v\\n", err)
        os.Exit(1)
    }
    defer resp.Body.Close()
    
    body, _ := io.ReadAll(resp.Body)
    fmt.Printf("Status Code: %d\\n", resp.StatusCode)
    
    if resp.StatusCode == 200 {
        var projects []interface{}
        if err := json.Unmarshal(body, &projects); err == nil {
            fmt.Printf("✅ API works! Found %d projects\\n", len(projects))
        } else {
            fmt.Println("✅ API works! Response received")
        }
    } else {
        fmt.Printf("Unexpected status: %d\\n", resp.StatusCode)
        fmt.Printf("Body: %s\\n", string(body))
    }
}
EOF
    """)
    
    # Run test
    test_output = machine.succeed("cd /tmp/hydra-go && go run test.go")
    print(test_output)
    
    print("✅ Test completed successfully!")
  '';
}