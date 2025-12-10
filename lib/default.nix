{ nixpkgs }:
{
  nixosSystem = args@{ system, modules, deployment ? {}, ... }:
    let
      certDir = let dir = builtins.getEnv "NIXOS_INCEPTION_CERT_DIR"; in
        if dir == ""
          then throw "NIXOS_INCEPTION_CERT_DIR not set"
        else dir;
      architectEndpoint = "${deployment.serverAddr or "127.0.0.1"}:${toString (deployment.serverPort or 8443)}";
      dreamer = nixpkgs.legacyPackages.${system}.buildGoModule {
        pname = "dreamer";
        version = "0.0.1";
        src = ../packages/dreamer;
        vendorHash = null;
      };
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
              "nixos-inception/ca.crt".source = builtins.path {
                path = "${certDir}/ca.crt";
                name = "ca.crt";
              };
              "nixos-inception/dreamer.crt".source = builtins.path {
                path = "${certDir}/dreamer.crt";
                name = "dreamer.crt";
              };
              "nixos-inception/dreamer.key" = {
                source = builtins.path {
                  path = "${certDir}/dreamer.key";
                  name = "dreamer.key";
                };
                mode = "0400";
              };
              "nixos-inception/config".text = architectEndpoint;
            };
            systemd.services.dreamer = {
              description = "NixOS Inception Dreamer";
              wantedBy = [ "multi-user.target" ];
              after = [ "network-online.target" ];
              wants = [ "network-online.target" ];
              serviceConfig = {
                Type = "oneshot";
                RemainAfterExit = true;
                ExecStart = "${dreamer}/bin/dreamer";
                Restart = "on-failure";
                RestartSec = "5s"; # TODO: Make configurable
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
