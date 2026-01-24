# Jumpboot Examples

Working examples demonstrating jumpboot features.

## Getting Started

Each example creates a Python environment in `examples/environments/`. The first run downloads micromamba and installs dependencies.

```bash
cd examples/repl
go run main.go
```

## Examples

| Example | Description |
|---------|-------------|
| [repl](repl/) | Interactive Python REPL from Go |
| [jsonqueueserver](jsonqueueserver/) | Bidirectional RPC with MessagePack |
| [jsonqueue](jsonqueue/) | Legacy JSON-based queue communication |
| [embedded_packages](embedded_packages/) | Complex package structures with go:embed |
| [chromadb](chromadb/) | Vector database with embeddings |
| [gradio](gradio/) | Web UI served from embedded Python |
| [mlx](mlx/) | Apple Silicon ML inference with MLX |
| [opencv_gui](opencv_gui/) | OpenCV image processing via Python |
| [tkinter](tkinter/) | Tkinter GUI from Go |

## Environment Cleanup

To remove all downloaded environments:

```bash
rm -rf examples/environments
```
