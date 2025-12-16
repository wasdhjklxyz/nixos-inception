{ nixpkgs }:
let
  lib = nixpkgs.lib;
  deploymentSchema = import ./deployment.nix { inherit lib; };
in {
  nixosSystem = args@{ system, modules, deployment ? {}, ... }:
    let
      deploy = deploymentSchema.validate deployment;
      certDir = let dir = builtins.getEnv "NIXOS_INCEPTION_CERT_DIR"; in
        if dir == "" then throw "NIXOS_INCEPTION_CERT_DIR not set" else dir;
      installerModule = import ./installer.nix {
        inherit nixpkgs system certDir deploy;
      };
      baseArgs = builtins.removeAttrs args [ "deployment" ];
      baseSystem = lib.nixosSystem baseArgs;
      _isoSystem = lib.nixosSystem {
        inherit system;
        modules = [
          (nixpkgs + "/nixos/modules/installer/cd-dvd/installation-cd-minimal.nix")
          installerModule
        ];
      };
    in baseSystem // {
      _inception = {
        iso = _isoSystem;
        deploymentConfig = deploy;
      };
    };
}
