{ pkgs, lib, ... }:
{
  home.extraActivationPath = [ pkgs.git ];
  home.activation = {
    setupRepoAction = lib.hm.dag.entryAfter [ "installPackages" ] ''
      if ! [[ -d "$HOME/.dotfiles" ]]; then
        $DRY_RUN_CMD git clone https://github.com/wkozyra95/dotfiles $HOME/.dotfiles
        $DRY_RUN_CMD pushd $HOME/.dotfiles
        $DRY_RUN_CMD git remote set-url origin git@github.com:wkozyra95/dotfiles.git
        $DRY_RUN_CMD popd
      fi
    '';
  };
}
