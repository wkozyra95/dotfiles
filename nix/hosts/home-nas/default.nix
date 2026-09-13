{ nixpkgs, overlays, inputs }:

let
  system = "x86_64-linux";
  custom = {
    unstable = import inputs.nixpkgs-unstable {
      inherit system;
      config = { allowUnfree = true; };
    };
    neovim-nightly = inputs.neovim-nightly-overlay.packages.${system}.default;
  };
in

nixpkgs.lib.nixosSystem {
  inherit system;

  specialArgs = { inherit custom; };

  modules = [
    inputs.home-manager.nixosModules.home-manager
    (import ../../nix-modules/myconfig.nix {
      username = "wojtek";
      email = "wkozyra95@gmail.com";
      env = "home-nas";
    })
    ./filesystems.nix
    ./system.nix
    ./boot.nix
    ./presence-suspend.nix
    ../../nix-modules/common.nix
    ({ config, lib, pkgs, ... }: {
      nixpkgs.overlays = overlays;
      home-manager = {
        extraSpecialArgs = {
          inherit custom;
        };
        useGlobalPkgs = true;
        useUserPackages = true;
        users.${config.myconfig.username} = (
          import ./home.nix config.myconfig.hm-modules
        );
      };

      services.k3s = {
        enable = true;
        role = "server";
        extraFlags = [
          "--disable=traefik"
          "--write-kubeconfig-mode=0644"
        ];
      };

      services.tailscale = {
        enable = true;
        useRoutingFeatures = "both";
        openFirewall = true;
      };

      networking.firewall.allowedTCPPorts = [ 6443 ];

      # Reports host metrics (cpu, memory, disks, temperatures) to the beszel
      # hub running in the cluster. It runs here rather than in a pod because a
      # container cannot see the real disks or sensors.
      #
      # KEY is the public key the hub shows when you add a system — it is not a
      # secret, it is how the agent recognises which hub may connect.
      systemd.services.beszel-agent = {
        description = "Beszel monitoring agent";
        after = [ "network-online.target" ];
        wants = [ "network-online.target" ];
        wantedBy = [ "multi-user.target" ];
        environment = {
          LISTEN = "45876";
          KEY = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIIv6x5V4vPE08xQgfBJGgk7zFlAku0ul51EIL4cy3l9B";
        };
        serviceConfig = {
          ExecStart = "${pkgs.beszel}/bin/beszel-agent";
          Restart = "on-failure";
          RestartSec = 5;
          StateDirectory = "beszel-agent";
          DynamicUser = true;
        };
      };

      # The hub connects in from a pod, so only the cluster bridge needs the
      # port open — not the LAN.
      networking.firewall.interfaces."cni0".allowedTCPPorts = [ 45876 ];

      environment.systemPackages = with pkgs; [
        usbutils
        fluxcd
        sops
        age
        kubectl
        kubernetes-helm
      ];

      users.users.${config.myconfig.username} = {
        extraGroups = [ "wireshark" ];
      };
    })
  ];
}
