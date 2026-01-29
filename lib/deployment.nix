{ lib }:
let
  deploymentOptions = {
    serverAddr = lib.mkOption {
      type = lib.types.str;
      default = "127.0.0.1";
      description = "Architect server address";
    };
    serverPort = lib.mkOption {
      type = lib.types.port;
      default = 8443;
      description = "Architect server port";
    };
    squashfsCompression = lib.mkOption {
      type = lib.types.str;
      default = "zstd -Xcompression-level 6";
      description = "Squashfs compression";
    };
    diskSelection = lib.mkOption {
      type = lib.types.enum [ "auto" "prompt" "specific" ];
      default = "specific";
      description = "Disk device selection type";
    };
    shipLock = lib.mkOption {
      type = lib.types.bool;
      default = true;
      description = "Ship flake.lock with deployment";
    };
    installerModule = lib.mkOption {
      type = lib.types.str;
      default = "/nixos/modules/installer/cd-dvd/installation-cd-minimal.nix";
    };
    bootOverrides = lib.mkOption {
      type = lib.types.nullOr lib.types.raw;
      default = null;
    };
  };
in {
  options = deploymentOptions;
  validate = deployment: (lib.evalModules {
    modules = [{
      options.d = deploymentOptions;
      config.d = deployment;
    }];
  }).config.d;
}
