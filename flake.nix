{
  description = "Zero-touch NixOS deployment with secrets management";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    systems.url = "github:nix-systems/default";
  };

  outputs = { self, nixpkgs, systems }:
    let
      eachSystem = nixpkgs.lib.genAttrs (import systems);
    in {
      lib = import ./lib { inherit nixpkgs; };
      packages = eachSystem (system: {
        architect = nixpkgs.legacyPackages.${system}.buildGoModule {
          pname = "architect";
          version = "0.0.1";
          src = ./packages/architect;
          vendorHash = "sha256-nYLz7vN7Cu8fUdJkaG3t0Qe3+k1+pGaPwlandN+YVmM=";
        };
      });
      apps = eachSystem (system: {
        default = {
          type = "app";
          program = "${self.packages.${system}.architect}/bin/architect";
        };
      });
    };
}
