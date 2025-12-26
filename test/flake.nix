{
  description = "Test flake for nixos-inception";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    disko = {
      url = "github:nix-community/disko";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    nixos-inception = {
      url = "path:../";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, disko, nixos-inception }: {
    nixosConfigurations.foo = nixos-inception.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        disko.nixosModules.disko
        ./disko.nix
        ./config.nix
      ];
      deployment = {
        ageKeyFile = "./secrets/key.txt";
        serverAddr = "10.0.2.2";
        serverPort = 12345;
        squashfsCompression = "zstd -Xcompression-level 1";
        diskSelection = "prompt";
      };
    };
  };
}
