{
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
      url = "github:wasdhjklxyz/nixos-inception";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, disko, sops-nix, nixos-inception }: {
    nixosConfigurations.rpi4 = nixos-inception.lib.nixosSystem {
      system = "aarch64-linux";
      modules = [
        disko.nixosModules.disko
        sops-nix.nixosModules.sops
        ./config.nix
      ];
      deployment.installerModule =
        "/nixos/modules/installer/sd-card/sd-image-aarch64-installer.nix";
    };
  };
}
