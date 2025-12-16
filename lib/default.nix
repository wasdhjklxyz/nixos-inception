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
      _netbootSystem = lib.nixosSystem {
        inherit system;
        modules = [
          (nixpkgs + "/nixos/modules/installer/netboot/netboot-minimal.nix")
          installerModule
        ];
      };
      _bootSystem = if deploy.bootMode == "netboot"
        then _netbootSystem else _isoSystem;
    in baseSystem // {
      _inception = {
        iso = _isoSystem;
        netboot = _netbootSystem;
        boot = _bootSystem;
        deploymentConfig = deploy;
      };
    };
}
