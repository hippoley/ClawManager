# edgefit

One command to answer a practical question: **what local AI model should I run on this machine, how, and why?**

edgefit is a standalone Go CLI staged inside ClawManager on an isolated feature branch. The code is intentionally self-contained under `tools/edgefit` so it can be extracted into its own repository without dragging ClawManager dependencies with it.

## Closed loop

hardware evidence
→ capacity pools
→ model + quant + context memory estimate
→ GPU / CPU-offload / CPU-only solver
→ runtime suggestion
→ task-aware ranking
→ executable command
→ real benchmark evidence

The important distinction is **estimate → executable plan → measured evidence**. A recommendation is not treated as proven until a real benchmark validates it on the machine.

## Commands

~~~bash
cd tools/edgefit
go test ./...
go build -o edgefit .

./edgefit scan
./edgefit doctor
./edgefit list
./edgefit recommend --task coding --top 5
./edgefit recommend --task reasoning --context 32768 --json
./edgefit plan --model qwen2.5-coder-7b-instruct --quant Q4_K_M --context 8192

# after loading a real model in Ollama
./edgefit bench --model qwen2.5-coder:7b
~~~

## What is implemented

- Linux/macOS/Windows CPU and RAM discovery
- NVIDIA VRAM/free-VRAM discovery through nvidia-smi
- Apple Silicon unified-memory path
- embedded model catalog
- quant-aware weight-memory estimate
- context-sensitive KV estimate
- runtime overhead allowance
- free-memory-aware fit grading
- GPU, CPU offload, and CPU-only paths
- task-aware ranking for general/coding/reasoning
- llama.cpp/Ollama, MLX, and vLLM-oriented runtime selection
- JSON output for automation
- runtime doctor checks
- measured Ollama tokens/sec and TTFT
- planner unit tests

## Why not just use total VRAM?

A nominal 24 GB GPU is not a 24 GB empty pool. edgefit prefers current free VRAM when the driver exposes it. That makes the answer closer to “will this launch now?” than “could this theoretically fit after rebooting and closing everything?”

## Evidence model

Each recommendation exposes:

- estimated weight memory
- estimated KV memory
- total required memory
- memory pool used for the decision
- remaining headroom
- execution path
- estimated throughput
- task score
- executable launch suggestion

The benchmark command then gives us a calibration target. Future device/model coefficients should be updated from measured error rather than hand-waved performance claims.

## Next gates for full parity+

1. **Catalog** — generated Hugging Face/provider ingestion, provenance, 100+ model variants.
2. **Hardware** — AMD/ROCm, Intel/SYCL, Windows GPU fallbacks, mixed/multi-GPU accounting.
3. **Planner** — architecture-aware KV cache, GQA/MQA, MoE expert residency, context and KV-quant alternatives.
4. **Runtime adapters** — Ollama, llama.cpp server, MLX, vLLM, LM Studio/OpenAI-compatible endpoints.
5. **Reality calibration** — real-machine benchmark matrix, estimate error and confidence intervals as first-class outputs.
6. **Distribution** — install script, releases, Homebrew/Scoop, container, and optional TUI/web UI.

This branch is an implementation spike with a runnable core, not a claim that the full hardware/model matrix has already been empirically proven.
