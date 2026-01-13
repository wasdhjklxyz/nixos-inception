{ nixpkgs }:
let
  lib = nixpkgs.lib;
  deploymentSchema = import ./deployment.nix { inherit lib; };
in {
  nixosSystem = args@{ system, modules, deployment ? {}, ... }:
    let
      deploy = deploymentSchema.validate deployment;
      buildSystem = let sys = builtins.getEnv "NIXOS_INCEPTION_BUILD_SYSTEM"; in
        if sys == "" then system else sys;
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
      sopsFile = builtins.toString baseSystem.config.sops.defaultSopsFile;
      sopsKeyPath = baseSystem.config.sops.age.keyFile;
      stateVersion = baseSystem.config.system.stateVersion;
      installerModule = import ./installer.nix {
        inherit nixpkgs system certDir deploy stateVersion;
      };
      _isoSystem = lib.nixosSystem {
        inherit system;
        modules = [
          (nixpkgs + "/nixos/modules/installer/cd-dvd/installation-cd-minimal.nix")
          installerModule
          { isoImage.squashfsCompression = deploy.squashfsCompression; }
        ];
      };
      _netbootSystem = lib.nixosSystem {
        inherit system;
        modules = [
          (nixpkgs + "/nixos/modules/installer/netboot/netboot-minimal.nix")
          installerModule
          { netboot.squashfsCompression = deploy.squashfsCompression; }
        ];
      };
      _bootSystem = if deploy.bootMode == "netboot"
        then _netbootSystem else _isoSystem;
    in baseSystem // {
      _inception = {
        inherit diskoDevice sopsFile sopsKeyPath;
        iso = _isoSystem;
        netboot = _netbootSystem;
        boot = _bootSystem;
        deploymentConfig = deploy;
      };
    };
}
