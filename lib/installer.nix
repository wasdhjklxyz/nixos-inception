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
  services.timesyncd.enable = true;
  systemd.services.systemd-time-wait-sync = {
    enable = true;
    wantedBy = [ "time-sync.target" ];
    serviceConfig.TimeoutStartSec = "30s";
  };
  systemd.services.set-approximate-time = {
    description = "Set approximate build time for RTC-less devices";
    wantedBy = [ "sysinit.target" ];
    before = [ "systemd-timesyncd.service" "time-sync.target" ];
    unitConfig = {
      ConditionPathExists = "!/run/clock-set";
      DefaultDependencies = false;
    };
    serviceConfig = {
      Type = "oneshot";
      ExecStart = "${pkgs.coreutils}/bin/date -s @${toString builtins.currentTime}";
      ExecStartPost = "${pkgs.coreutils}/bin/touch /run/clock-set";
    };
  };
  systemd.services.dreamer =
    let
      deps = [
        "network-online.target"
        "sshd-keygen.service"
        "time-sync.target"
      ];
    in {
      description = "NixOS Inception Dreamer";
      wantedBy = [ "multi-user.target" ];
      after = deps;
      wants = deps;
      serviceConfig = {
        Type = "oneshot";
        ExecStart = "${dreamer}/bin/dreamer";
        ExecStartPost = "${pkgs.systemd}/bin/systemctl reboot";
        PrivateTmp = false;
      };
      path = with pkgs; [ nix util-linux nixos-install-tools coreutils ];
    };
  system.stateVersion = stateVersion;
}
