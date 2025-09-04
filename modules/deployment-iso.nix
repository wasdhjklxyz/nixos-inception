{ config, pkgs, modulesPath, ... }:
{
  imports = [ (modulesPath + "/installer/cd-dvd/installation-cd-minimal.nix") ];

  assertions = [
    {
      assertion = config.disko.devices != {};
      message = "Configuration must include disko partitioning scheme";
    }
  ];

  environment.systemPackages = with pkgs; [ curl ];

  systemd.services.deployment-beacon = {
    wantedBy = [ "multi-user.target" ];
    after = [ "network-online.target" ];
    script = ''
    '';
  };
}
