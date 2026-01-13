{
  description = "Test flake for nixos-inception";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    disko = {
      url = "github:nix-community/disko";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    sops-nix = {
      url = "github:Mic92/sops-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    nixos-inception = {
      url = "path:../";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, disko, sops-nix, nixos-inception }: {
    nixosConfigurations.foo = nixos-inception.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        disko.nixosModules.disko
        ./disko.nix
        sops-nix.nixosModules.sops
        ./config.nix
      ];
      deployment = {
        serverAddr = "10.0.2.2";
        serverPort = 12345;
        squashfsCompression = "zstd -Xcompression-level 1";
        diskSelection = "prompt";
      };
    };
  };
}
