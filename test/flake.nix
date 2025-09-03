{
  description = "Test flake for nixos-inception";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
  };

  outputs = { self, nixpkgs }: {
    nixosConfigurations.test = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        {
          boot.loader.grub.device = "/dev/null";
          fileSystems."/" = { device = "/dev/null"; fsType = "ext4"; };
          system.stateVersion = "25.05";
        }
      ];
    };
  };
}
