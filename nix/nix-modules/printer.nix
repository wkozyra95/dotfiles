{ pkgs, config, ... }:
{
  services.printing.enable = true;
  services.printing.drivers = [
    pkgs.brlaser
    pkgs.brgenml1lpr
    pkgs.brgenml1cupswrapper
  ];
  services.avahi = {
    enable = true;
    openFirewall = true;
  };
  services.saned.enable = true;
  hardware.sane = {
    enable = true;
    brscan4.enable = true;
    extraBackends = [
      pkgs.brscan4
    ];
  };

  users.users.${config.myconfig.username} = {
    extraGroups = [ "scanner" "lp" ];
  };
  services.ipp-usb.enable = true;

  myconfig.hm-modules = [
    {
      home.packages = [
        pkgs.simple-scan
      ];
    }
  ];
}
