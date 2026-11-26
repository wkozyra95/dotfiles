{ pkgs
  # Default speech speed multiplier; CLAUDE_TTS_SPEED overrides it at runtime.
, speed ? "2"
, ...
}:
let
  # Piper voice (https://huggingface.co/rhasspy/piper-voices). To swap voices,
  # change the path and both hashes (`nix store prefetch-file <url>`).
  voiceBase = "https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/en/en_US/hfc_male/medium/en_US-hfc_male-medium";
  voiceModel = pkgs.fetchurl {
    url = "${voiceBase}.onnx";
    hash = "sha256-0R5AOgK99aZwyHez3Fbg4cjOzm+zAolYYxTf/cCnjLA=";
  };
  voiceConfig = pkgs.fetchurl {
    url = "${voiceBase}.onnx.json";
    hash = "sha256-9mhHQkrtC/mey7XXz95HwKkG9Cag2vfEbzBefSGv2IY=";
  };
  # piper 1.4 parses --config but never uses it: the config must sit at
  # "<model>.json", so put both files side by side.
  voice = pkgs.linkFarm "piper-voice" [
    { name = "voice.onnx"; path = voiceModel; }
    { name = "voice.onnx.json"; path = voiceConfig; }
  ];
in
# Reads Claude Code replies aloud. Wired up as hooks in ~/.claude/settings.json:
#   Stop             -> claude-tts hook-stop    (speak the reply, if enabled)
#   UserPromptSubmit -> claude-tts hook-prompt  (cut off the previous reply)
# Enabled per session, either at launch (CLAUDE_TTS=1, the `claude-voice` shell
# alias) or mid-session with the /tts skill, which runs
# `claude-tts toggle <session id>`. A per-session toggle overrides CLAUDE_TTS.
pkgs.writeShellApplication {
  name = "claude-tts";
  runtimeInputs = with pkgs; [ piper-tts pipewire jq gnused coreutils util-linux ];
  # Single-quoted sed scripts and the `bash -c` body are meant to stay literal.
  excludeShellChecks = [ "SC2016" ];
  text = ''
    state_dir="''${XDG_RUNTIME_DIR:-/tmp}/claude-tts"
    mkdir -p "$state_dir"

    # on/off for a session: explicit toggle wins, else CLAUDE_TTS from launch.
    enabled() {
      local state_file="$state_dir/$1.state"
      if [[ -f "$state_file" ]]; then
        [[ "$(cat "$state_file")" == on ]]
      else
        [[ -n "''${CLAUDE_TTS:-}" ]]
      fi
    }

    # Playback runs as its own process group (setsid), pgid stored in a file.
    stop_playback() {
      local pid_file="$state_dir/$1.pid"
      [[ -f "$pid_file" ]] || return 0
      kill -- "-$(cat "$pid_file")" 2>/dev/null || true
      rm -f "$pid_file"
    }

    # Fed to Claude (UserPromptSubmit stdout becomes context) on every turn
    # while speech is on, so it holds for claude-voice launches and survives
    # compaction. Matches what clean() below can and can't speak.
    guidance() {
      cat <<'EOF'
    Speech output is ON for this session: each reply is read aloud by a text-to-speech engine. Fenced code blocks and tables are skipped entirely, and inline code that is not a plain word is dropped. Write replies for listening:
    - Plain conversational prose in short sentences. No tables, headings or nested lists; a short flat list is fine.
    - Keep exact strings (commands, code, file paths, URLs) out of sentences. Put them in fenced code blocks, which are shown on screen but not spoken, and say in prose what they are or do.
    - Lead with the answer, keep it brief, and avoid symbols and abbreviations that sound odd when read aloud.
    EOF
    }

    # Markdown -> speakable prose: drop code blocks, tables, URLs and markup.
    # Inline code is kept only when it is plain words (`git status`, `jq`);
    # anything with paths or symbols is dropped rather than spelled out.
    clean() {
      sed -E \
        -e '/^[[:space:]]*(```|~~~)/,/^[[:space:]]*(```|~~~)/d' \
        -e 's/`([A-Za-z0-9 ._-]+)`/\1/g' \
        -e 's/`[^`]*`//g' \
        -e '/^[[:space:]]*\|/d' \
        -e '/^[[:space:]]*[-*_]{3,}[[:space:]]*$/d' \
        -e 's/\[([^]]*)\]\([^)]*\)/\1/g' \
        -e 's#https?://[^[:space:])]+##g' \
        -e 's/^[[:space:]]*#+[[:space:]]*//' \
        -e 's/^[[:space:]]*>[[:space:]]?//' \
        -e 's/^[[:space:]]*([-*+]|[0-9]+\.)[[:space:]]+//' \
        -e 's/[`*]//g' \
        -e 's/(^|[[:space:]])_+|_+([[:space:]]|$)/\1\2/g'
    }

    speak() {
      local session="$1" text="$2"
      stop_playback "$session"
      [[ -n "''${text//[[:space:]]/}" ]] || return 0
      # Speed multiplier (CLAUDE_TTS_SPEED, default ${speed}): piper's knob is
      # phoneme length, so divide the voice's default length_scale by it.
      local length_scale
      length_scale="$(jq -n --arg s "''${CLAUDE_TTS_SPEED:-${speed}}" \
        --slurpfile c "${voiceConfig}" \
        '($c[0].inference.length_scale // 1) / ($s | tonumber)' 2>/dev/null)" \
        || length_scale=1
      # Loudness is left to PipeWire, not processed here (piper already peaks at
      # full scale; a limiter audibly degraded the voice). pw-play defaults to
      # media.role=Music, and WirePlumber restores one saved volume per role, so
      # TTS inherited whatever Music was last set to. Accessibility (the screen
      # reader role) gives it a volume of its own: adjust it in pavucontrol
      # while it speaks ("Claude TTS") and WirePlumber remembers it.
      # Not a process-group leader here, so setsid execs in place: $! is the pgid.
      # All fds detached so the hook returns immediately.
      setsid bash -c '
        piper -m "$1" --length-scale "$3" --output-raw 2>/dev/null \
          | pw-play --raw --rate "$2" --channels 1 --format s16 \
              --media-role Accessibility \
              -P "{ application.name = \"Claude TTS\" media.name = \"Claude reply\" }" -
      ' _ "${voice}/voice.onnx" \
        "$(jq -r .audio.sample_rate "${voiceConfig}")" "$length_scale" \
        <<<"$text" >/dev/null 2>&1 &
      echo $! >"$state_dir/$session.pid"
    }

    cmd="''${1:-}"
    case "$cmd" in
      hook-stop)
        input="$(cat)"
        session="$(jq -r .session_id <<<"$input")"
        enabled "$session" || exit 0
        # Newer Claude Code passes the reply directly; fall back to the transcript.
        text="$(jq -r '.last_assistant_message // empty' <<<"$input")"
        if [[ -z "$text" ]]; then
          transcript="$(jq -r .transcript_path <<<"$input")"
          text="$(jq -rs '
            [.[] | select(.type == "assistant")] | last
            | [.message.content[]? | select(.type == "text") | .text] | join("\n")
          ' "$transcript")"
        fi
        speak "$session" "$(clean <<<"$text")"
        ;;
      hook-prompt)
        session="$(jq -r .session_id)"
        stop_playback "$session"
        if enabled "$session"; then guidance; fi
        ;;
      toggle)
        session="''${2:?usage: claude-tts toggle <session-id>}"
        if enabled "$session"; then
          echo off >"$state_dir/$session.state"
          stop_playback "$session"
          echo "Speech OFF for this session. Replies are no longer read aloud; normal formatting (tables, code, paths) is fine again."
        else
          echo on >"$state_dir/$session.state"
          guidance
        fi
        ;;
      stop)
        # Silence one session, or every session when none is given.
        if [[ -n "''${2:-}" ]]; then
          stop_playback "$2"
        else
          for f in "$state_dir"/*.pid; do
            if [[ -e "$f" ]]; then stop_playback "$(basename "$f" .pid)"; fi
          done
        fi
        ;;
      say)
        # Manual test: echo "hello" | claude-tts say
        speak manual "$(clean)"
        ;;
      *)
        echo "usage: claude-tts {hook-stop|hook-prompt|toggle <session-id>|stop [session-id]|say}" >&2
        exit 1
        ;;
    esac
  '';
}
