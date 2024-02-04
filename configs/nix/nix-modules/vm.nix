{ pkgs, ... }:
{
  # It might be necessary to run `sudo virsh net-autostart default` once
  # or `sudo virsh net-start default` on system startup

  virtualisation.libvirtd.enable = true;
  programs.virt-manager.enable = true;
  environment.systemPackages = with pkgs; [ qemu ];
  myconfig.hm-modules = [
    {
      dconf.settings = {
        "org/virt-manager/virt-manager/connections" = {
          autoconnect = [ "qemu:///system" ];
          uris = [ "qemu:///system" ];
        };
      };
    }
  ];
}
