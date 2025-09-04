{ lib, ... }:
{
  options.deployment = {
    ageKeyFile = lib.mkOption {
      type = lib.types.nullOr lib.types.path;
      default = null;
      description = "Path to age identity file";
    };
    serverPort = lib.mkOption {
      type = lib.types.int;
      default = 12345;  # TODO: architect/config.go looks up http service port
                        #       these should be matching or decide to use one
                        #       thats hardcoded
      description = "Listen port";
    };
  };
}
