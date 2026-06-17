{ custom, ... }:
{
  # Ollama backend for agentic harnesses (opencode). Ollama parses tool-calls
  # itself (in Go), so opencode's tool-calls execute reliably.
  #
  # Uses the *Vulkan* build: ROCm in nixpkgs is too old to enumerate this
  # RDNA4 / gfx1201 card (falls back to CPU, which is too slow), but Vulkan
  # works great on RADV.
  #
  # Listens on 127.0.0.1:11434 (OpenAI-compatible API at /v1).
  services.ollama = {
    enable = true;
    package = custom.unstable.ollama-vulkan;
    # Official library qwen3-coder:30b (~18 GB Q4_K_M). Larger than VRAM so a
    # few layers spill to CPU, but the library build ships Ollama's correct
    # tool-call template/parser, so agentic tool-calls in opencode actually
    # execute. The 3B-active MoE keeps the CPU spill bearable.
    loadModels = [ "qwen3-coder:30b" ];
    environmentVariables = {
      # Vulkan enumerates two devices: 0 = discrete RX 9070 XT (gfx1201),
      # 1 = the CPU's integrated GPU (RAPHAEL, uma, slow). Pin to ONLY the
      # discrete card so Ollama doesn't split the model onto the iGPU.
      GGML_VK_VISIBLE_DEVICES = "0";
      # Flash attention shrinks the KV cache a lot and speeds up generation.
      OLLAMA_FLASH_ATTENTION = "1";
      # Quantize the KV cache (halves its VRAM) so more layers fit on the GPU.
      OLLAMA_KV_CACHE_TYPE = "q8_0";
      # Keep context sane (the model's native default is huge -> giant KV alloc).
      # 24k + q8_0 KV keeps total VRAM under the card's 15.9 GB.
      OLLAMA_CONTEXT_LENGTH = "24576";
    };
  };
}
