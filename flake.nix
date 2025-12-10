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
          src = ./architect;
          vendorHash = "sha256-Jc8biA1JZkvcA/kXjE/9MCn6CftRlmb4G5x6MHYeVMA=";
        };
      });
      apps = eachSystem (system: {
        default = {
          type = "app";
          program = "${nixpkgs.legacyPackages.${system}.writeShellScript
            "nixos-inception" ''
            export PATH=${nixpkgs.legacyPackages.${system}.lib.makeBinPath [
              self.packages.${system}.architect
              nixpkgs.legacyPackages.${system}.jq
            ]}:$PATH
            ${builtins.readFile ./scripts/nixos-inception.sh}
          ''}";
        };
      });
    };
}
