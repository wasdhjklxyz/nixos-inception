{ ... }:
{
  boot.loader.grub.device = "/dev/null";
  fileSystems."/" = { device = "/dev/null"; fsType = "ext4"; };
  system.stateVersion = "25.05";
}
