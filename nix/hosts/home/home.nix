systemModules:
{ pkgs, config, ... }:
{
  imports = [
    ../../hm-modules/common.nix
    ../../hm-modules/common-desktop.nix
    ../../hm-modules/git.nix
    ../../hm-modules/vim.nix
    ../../hm-modules/neovim.nix
    ../../hm-modules/dotfiles.nix
    ../../hm-modules/dunst.nix
  ] ++ systemModules;

  config = {
    home.packages = with pkgs; [
      bitwarden-cli
      gh
      obs-studio
      prusa-slicer
      ueberzugpp
      lutris
      wine
      opencode # terminal coding agent; talks to the local Ollama server
    ];

    # Launcher entry ($mod+p) for `mycli mobile send`: fuzzel asks for the device
    # and the text, the push goes to the phones registered with hostd. Only works
    # here, hostd runs as the session user on this host.
    xdg.desktopEntries.myremote-send = {
      name = "MyRemote Notification";
      comment = "Push text to the phone";
      exec = "mycli mobile send";
      terminal = false;
      icon = "phone";
      categories = [ "Utility" ];
    };

    # Live-editable (like the nvim config); points at the local Ollama server.
    xdg.configFile."opencode/opencode.json".source =
      config.lib.file.mkOutOfStoreSymlink
        "${config.home.homeDirectory}/.dotfiles/configs/opencode/opencode.json";

    myconfig = {
      git.signingKey = "35DF 8DFA D0E7 1E39 F047 BD01 AE51 A568 2B78 648C";
    };

    home.sessionVariables = {
      CARGO_BUILD_JOBS = "16";
    };

    programs.gpg.enable = true;
    services.gpg-agent = {
      enable = true;
      pinentry = {
        package = pkgs.pinentry-curses;
      };
      enableSshSupport = true;
      enableExtraSocket = true;
    };

    programs.ranger.extraConfig = ''
      set preview_images true
      set preview_images_method ueberzug
    '';

    programs.home-manager.enable = true;
    home.stateVersion = "23.11";
  };
}
