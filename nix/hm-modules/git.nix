{ config, lib, ... }:
let
  cfg = config.myconfig;
in
{
  options.myconfig.git = {
    signingKey = lib.mkOption {
      type = lib.types.nullOr lib.types.str;
      default = null;
    };
  };
  config = {
    programs.git = {
      enable = true;
      ignores = [
        "compile_commands.json"
        "**/.cache/clangd/**"
        ".git"
      ];
      settings = {
        user = {
          email = cfg.email;
          name = "Wojciech Kozyra";
        };
        core.editor = "nvim";
        status.submoduleSummary = true;
        init.defaultBranch = "main";
        pull.rebase = true;
        push.default = "current";
        merge.conflictStyle = "diff3";
      };
      lfs.enable = true;
      signing = lib.mkIf (cfg.git.signingKey != null) {
        key = cfg.git.signingKey;
        signByDefault = true;
      };
    };
    programs.diff-so-fancy = {
      enable = true;
      enableGitIntegration = true;
      settings = {
        markEmptyLines = false;
      };
      pagerOpts = [ "--tabs=4" "-RF" ];
    };
  };
}
