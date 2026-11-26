---
name: tts
description: Toggle reading Claude's replies aloud (piper TTS) for the current session only.
disable-model-invocation: true
allowed-tools: Bash(claude-tts toggle:*)
---

<!-- Toggle runs at skill load; the Stop/UserPromptSubmit hooks in
     ~/.claude/settings.json do the speaking and repeat the formatting
     guidance each turn while speech is on (nix/packages/claude-tts.nix). -->

!`claude-tts toggle ${CLAUDE_SESSION_ID}`

The output above is the new speech state for this session. Confirm it to the user in one short sentence, and from now on format replies as it says. Do nothing else.
