{ ... }:
{
  disko.devices.disk.main = {
    # Required by disko. If deployment.diskSelection="specific", this device is
    # used as-is. Otherwise it's a placeholder and architect selects the disk.
    device = "/dev/disk/by-id/foobarbaz";
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
}
