{ config, ... }:
{
  disko.devices.disk.main = {
    device = "/dev/disk/by-id/nixos-inception-placeholder";
    type = "disk";
    content = {
      type = "gpt";
      partitions = {
        MBR = {
          type = "EF02";
          size = "1M";
          priority = 1;
        };
        ESP = {
          type = "EF00";
          size = "500M";
          content = {
            type = "filesystem";
            format = "vfat";
            mountpoint = "/boot";
            mountOptions = [ "umask=0077" ];
          };
        };
        root = {
          size = "100%";
          content = {
            type = "filesystem";
            format = "ext4";
            mountpoint = "/";
          };
        };
      };
    };
  };

  sops = {
    defaultSopsFile = ./secrets/secrets.yaml;
    age.keyFile = "/var/lib/sops-nix/key.txt";
    secrets.password.neededForUsers = true;
  };

  users.users.user = {
    isNormalUser = true;
    hashedPasswordFile = config.sops.secrets.password.path;
    extraGroups = [ "wheel" ];
  };

  system.stateVersion = "25.11";
}
