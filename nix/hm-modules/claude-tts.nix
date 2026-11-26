{ pkgs, config, lib, ... }:
# Claude Code replies read aloud (piper, nix/packages/claude-tts.nix). Off by
# default; per session either launch with `claude-voice` or type /tts inside a
# running session.
let
  claude-tts = pkgs.callPackage ../packages/claude-tts.nix { };
  settings = "${config.home.homeDirectory}/.claude/settings.json";
in
{
  home.packages = [ claude-tts ];

  programs.zsh.shellAliases.claude-voice = "CLAUDE_TTS=1 claude";

  home.file.".claude/skills/tts".source =
    config.lib.file.mkOutOfStoreSymlink
      "${config.home.homeDirectory}/.dotfiles/configs/claude/skills/tts";

  # ~/.claude/settings.json is rewritten by Claude Code itself, so it can't be a
  # store symlink. Instead merge the two hooks in on activation, only when a
  # hook with that command isn't there yet; everything else in the file is kept.
  home.activation.claudeTtsHooks = lib.hm.dag.entryAfter [ "writeBoundary" ] ''
    settings="${settings}"
    if [[ ! -f "$settings" ]]; then
      $DRY_RUN_CMD mkdir -p "$(dirname "$settings")"
      $DRY_RUN_CMD sh -c 'echo "{}" > "$1"' _ "$settings"
    fi
    if [[ -z "''${DRY_RUN:-}" ]]; then
      ${pkgs.jq}/bin/jq '
        def ensure($event; $cmd):
          if [.hooks[$event][]?.hooks[]?.command] | index($cmd) then .
          else .hooks[$event] += [{hooks: [{type: "command", command: $cmd}]}]
          end;
        ensure("Stop"; "claude-tts hook-stop")
        | ensure("UserPromptSubmit"; "claude-tts hook-prompt")
      ' "$settings" > "$settings.tmp"
      if ${pkgs.diffutils}/bin/cmp -s "$settings" "$settings.tmp"; then
        rm -f "$settings.tmp"
      else
        mv -f "$settings.tmp" "$settings"
        echo "claude-tts: added hooks to $settings"
      fi
    fi
  '';
}
