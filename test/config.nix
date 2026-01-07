{ ... }:
{
  sops = {
    defaultSopsFile = ./secrets.yaml;
    keyFile = ./secrets/key.txt;
  };
  system.stateVersion = "25.05";
}
