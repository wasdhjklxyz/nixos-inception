{ config, ... }:
{
  sops = {
    defaultSopsFile = ./secrets.yaml;
    # NOTE: Only supports age.keyFile (for now) a key is generated/written here.
    #       See https://github.com/wasdhjklxyz/nixos-inception/issues/24
    age.keyFile = "/var/lib/sops-nix/key.txt";
    secrets.foo-password.neededForUsers = true;
  };

  users.users.foo = {
    isNormalUser = true;
    hashedPasswordFile = config.sops.secrets.foo-password.path;
  };

  system.stateVersion = "25.05";
}
