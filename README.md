# dotfiles

## Install

On first install might be necessary to add
```
--extra-experimental-features "nix-command flakes"
```

#### Home - NixOS

```
sudo nixos-rebuild switch --flake ".#home"
```

#### Work - arch

```
nix run home-manager/release-24.11 -- switch --flake ".#work"
```

#### Work - macOS

```
nix run nix-darwin -- switch --flake ".#work-mac" 
```
