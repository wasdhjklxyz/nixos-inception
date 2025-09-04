{ nixpkgs }:
{
  nixosSystem = args@{ system, modules, deployment ? {}, ... }:
    let
      baseArgs = builtins.removeAttrs args [ "deployment" ];
      baseSystem = nixpkgs.lib.nixosSystem baseArgs;
      _isoSystem = nixpkgs.lib.nixosSystem {
        inherit system;
        modules = [
          (nixpkgs +
            "/nixos/modules/installer/cd-dvd/installation-cd-minimal.nix")
          {
            networking.hostName = "something";
          }
        ];
      };
    in baseSystem // {
      _inception = {
        iso = _isoSystem;
        deploymentConfig = deployment;
      };
    };
}
