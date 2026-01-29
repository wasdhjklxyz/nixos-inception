{ config, ... }:
{
  disko.devices.disk.main = {
    # When deployment.diskSelection is "auto" or "prompt", this MUST be set to
    # exactly "/dev/disk/by-id/nixos-inception-placeholder". The actual target
    # device is selected at install time by the architect.
    #
    # When deployment.diskSelection is "specific", set this to the actual
    # device path (e.g. "/dev/sda" or "/dev/disk/by-id/...").
    #
    # If you think this is cringe, I agree! See the following to discuss:
    #   https://github.com/wasdhjklxyz/nixos-inception/issues/19
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
    defaultSopsFile = ./secrets.yaml;
    # Only supports age.keyFile (for now) a key is generated/written here.
    #   https://github.com/wasdhjklxyz/nixos-inception/issues/24
    age.keyFile = "/var/lib/sops-nix/key.txt";
    secrets.password.neededForUsers = true;
  };

  users.users.user = {
    isNormalUser = true;
    hashedPasswordFile = config.sops.secrets.password.path;
    extraGroups = [ "wheel" ];
  };

  services.openssh.enable = true;
  system.stateVersion = "25.11";
}
