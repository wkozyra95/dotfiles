{ pkgs, config, ... }:
{
  # It might be necessary to run `sudo virsh net-autostart default` once
  # or `sudo virsh net-start default` on system startup

  virtualisation.libvirtd = {
    enable = true;
    qemu = {
      package = pkgs.qemu_kvm;
      runAsRoot = true;
      swtpm.enable = true;
    };
  };

  programs.virt-manager.enable = true;
  environment.systemPackages = with pkgs; [ qemu ];
  users.users.${config.myconfig.username} = {
    extraGroups = [ "libvirtd" "kvm" "qemu-libvirtd" ];
  };
  myconfig.hm-modules = [
    {
      home.pointerCursor = {
        gtk.enable = true;
        name = "Vanilla-DMZ";
        package = pkgs.vanilla-dmz;
      };
      dconf.settings = {
        "org/virt-manager/virt-manager/connections" = {
          autoconnect = [ "qemu:///system" ];
          uris = [ "qemu:///system" ];
        };
      };
    }
  ];
}
