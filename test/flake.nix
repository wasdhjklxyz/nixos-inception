{
  description = "Test flake for nixos-inception";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    nixos-inception = {
      url = "path:../";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, nixos-inception }: {
    nixosConfigurations.foo = nixos-inception.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [ ./foo.nix ];
      deployment = {
        ageKeyFile = "./secrets/key.txt";
        serverAddr = "10.0.2.2";
        serverPort = 12345;
      };
    };
  };
}
