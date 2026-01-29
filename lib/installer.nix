{ nixpkgs, system, certDir, deploy, stateVersion }:
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
  systemd.services.dreamer =
    let
      requires = [
        "network-online.target"
        "sshd-keygen.service"
        "time-sync.target"
      ];
    in {
      description = "NixOS Inception Dreamer";
      wantedBy = [ "multi-user.target" ];
      after = requires;
      wants = requires;
      serviceConfig = {
        Type = "oneshot";
        RemainAfterExit = true;
        ExecStart = "${dreamer}/bin/dreamer";
      };
      path = with pkgs; [ nix util-linux nixos-install-tools ];
    };
  services.timesyncd.enable = true;
  system.stateVersion = stateVersion;
}
