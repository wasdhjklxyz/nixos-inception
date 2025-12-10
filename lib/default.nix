{ nixpkgs }:
{
  nixosSystem = args@{ system, modules, deployment ? {}, ... }:
    let
      certDir = let dir = builtins.getEnv "NIXOS_INCEPTION_CERT_DIR"; in
        if dir == ""
          then throw "NIXOS_INCEPTION_CERT_DIR not set"
        else dir;
      baseArgs = builtins.removeAttrs args [ "deployment" ];
      baseSystem = nixpkgs.lib.nixosSystem baseArgs;
      _isoSystem = nixpkgs.lib.nixosSystem {
        inherit system;
        modules = [
          (nixpkgs +
            "/nixos/modules/installer/cd-dvd/installation-cd-minimal.nix")
          {
            networking.hostName = "something";
            environment.etc = {
              "nixos-inception/client.crt".source = builtins.path {
                path = "${certDir}/client.crt";
                name = "client.crt";
              };
              "nixos-inception/client.key" = {
                source = builtins.path {
                  path = "${certDir}/client.key";
                  name = "client.key";
                };
                mode = "0400";
              };
            };
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
