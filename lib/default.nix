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
      baseArgs = builtins.removeAttrs args [ "deployment" ];
      baseSystem = lib.nixosSystem baseArgs;
      _ = if !(baseSystem.config ? disko)
        then throw "nixos-inception requires disko module - add disko to your flake inputs and modules"
        else if baseSystem.config.disko.devices == {}
        then throw "nixos-inception requires disko.devices to be configured"
        else null;
      diskoDevice =
        let
          disks = baseSystem.config.disko.devices.disk;
          diskNames = builtins.attrNames disks;
        in
          if builtins.length diskNames != 1
          then throw ''
            nixos-inception currently supports single-disk configurations only.
            Multi-disk support: https://github.com/wasdhjklxyz/nixos-inception/issues/17
          ''
          else disks.${builtins.head diskNames}.device;
      stateVersion = baseSystem.config.system.stateVersion;
      installerModule = import ./installer.nix {
        inherit nixpkgs system certDir deploy stateVersion;
      };
      _bootSystem = lib.nixosSystem {
        inherit system;
        modules = [
          (nixpkgs + deploy.installerModule)
          installerModule
        ] ++ (if deploy.bootOverrides != null
          then [ deploy.bootOverrides ]
          else []
        );
      };
    in baseSystem // {
      _inception = {
        inherit diskoDevice;
        boot = _bootSystem;
        deploymentConfig = deploy;
      };
    };
}
