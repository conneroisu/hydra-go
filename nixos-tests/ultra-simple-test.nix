# Ultra-simple NixOS VM test with real Hydra
{ pkgs ? import <nixpkgs> {} }:

pkgs.nixosTest {
  name = "hydra-sdk-test";
  
  nodes.machine = { config, pkgs, ... }: {
    virtualisation.memorySize = 4096;
    
    services.postgresql.enable = true;
    
    services.hydra = {
      enable = true;
      hydraURL = "http://localhost:3000";
      notificationSender = "test@localhost";
    };
    
    nix.settings.sandbox = false;
    
    environment.systemPackages = [ pkgs.curl ];
  };
  
  testScript = ''
    machine.start()
    machine.wait_for_unit("multi-user.target")
    machine.wait_for_unit("postgresql.service")
    machine.wait_for_unit("hydra-init.service")
    machine.wait_for_unit("hydra-server.service")
    machine.wait_for_open_port(3000)
    machine.succeed("curl http://localhost:3000/")
    print("Hydra is running!")
  '';
}