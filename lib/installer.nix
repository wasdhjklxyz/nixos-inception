{ certDir, deploy, stateVersion }:
{ config, pkgs, lib, ... }:
let
  pkgs = nixpkgs.legacyPackages.${system};
  architectEndpoint = "${deploy.serverAddr}:${toString deploy.serverPort}";
  dreamer = pkgs.buildGoModule {
    pname = "dreamer";
    version = "0.0.1";
    src = ../packages/dreamer;
    vendorHash = "sha256-3V8KBFJKjZ/9aE5dzFEzWK+TU+3uhcdwPzC9ANmnBGA=";
  };
in {
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
    after = [ "network-online.target" "sshd-keygen.service" ];
    wants = [ "network-online.target" "sshd-keygen.service" ];
    serviceConfig = {
      Type = "oneshot";
      RemainAfterExit = true;
      ExecStart = "${dreamer}/bin/dreamer";
    };
    path = with pkgs; [ nix util-linux nixos-install-tools ];
  };
  system.stateVersion = stateVersion;
}
