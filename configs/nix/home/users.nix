{ pkgs, ...}:

{
  users = {
    mutableUsers = true;
    users = {
      wojtek = {
        isNormalUser = true;
        home = "/home/wojtek";
        description = "Wojtek";
        extraGroups = [ "wheel" "networkmanager" "adbusers" "libvirtd" "docker" "kvm" "qemu-libvirtd" "sudo" "audio"];
        shell = pkgs.zsh;
      };

      root = {
        home = "/root";
      };
    };
  };
}
