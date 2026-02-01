{ pkgs, ... }:
{
  services = {
    openssh.enable = true;
    getty.autologinUser = "root";
  };
  environment.systemPackages = with pkgs; [ vim ];
}
