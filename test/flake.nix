{
  description = "Test flake for nixos-inception";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    nixos-inception = {
      url = "path:../";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, nixos-inception }:
    let
      pkgs = import nixpkgs {
        overlays = [ nixos-inception.overlays.default ];
      };
    in {
    nixosConfigurations.foo = pkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        ./foo.nix
        {
          deployment.ageKeyFile = ./secrets/key.txt;
          deployment.serverPort = 12345;
        }
      ];
    };
  };
}
