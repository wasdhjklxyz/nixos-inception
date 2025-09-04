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
      packages = eachSystem (system: {
        architect = nixpkgs.legacyPackages.${system}.buildGoModule {
          pname = "architect";
          version = "0.0.1";
          src = ./architect;
          vendorHash = "sha256-Jc8biA1JZkvcA/kXjE/9MCn6CftRlmb4G5x6MHYeVMA=";
        };
      });

      overlays.default = final: prev: {
        lib = prev.lib // {
          nixosSystem = args:
            let
              base = prev.lib.nixosSystem args;
            in base // {
              deployable = base.extendModules {
                modules = [ ];
              };
            };
        };
      };
    };
}
