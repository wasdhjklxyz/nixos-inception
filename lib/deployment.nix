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
    bootMode = lib.mkOption {
      type = lib.types.enum [ "iso" "netboot" ];
      default = "iso";
      description = "Boot medium type";
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
