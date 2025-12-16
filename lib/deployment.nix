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
    ageKeyFile = lib.mkOption {
      type = lib.types.nullOr lib.types.str;
      default = null;
      description = "Path to age identity file";
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
