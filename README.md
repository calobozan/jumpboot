# Jumpboot

[![Go Reference](https://pkg.go.dev/badge/github.com/richinsley/jumpboot.svg)](https://pkg.go.dev/github.com/richinsley/jumpboot)
[![Go Report Card](https://goreportcard.com/badge/github.com/richinsley/jumpboot)](https://goreportcard.com/report/github.com/richinsley/jumpboot)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Embed Python and its entire ecosystem in your Go binary. Ship a single executable that creates its own Python environment on first run.

```go
//go:embed sentiment.py
var sentimentCode string

func main() {
    // Creates isolated Python 3.11 environment with pip (first run only)
    env, _ := jumpboot.CreateEnvironmentMamba("mlenv", "./envs", "3.11", "conda-forge", nil)

    if env.IsNew {
        env.PipInstallPackages([]string{"transformers", "torch"}, "", "", false, nil)
    }

    // Embed Python code in your binary, call it like a function
    mod := jumpboot.NewModuleFromString("sentiment", "sentiment.py", sentimentCode)
    repl, _ := env.NewREPLPythonProcess(nil, nil, []jumpboot.Module{*mod}, nil)
    defer repl.Close()

    repl.Execute("import sentiment", true)
    result, _ := repl.Execute(`sentiment.analyze("I love this product!")`, true)
    fmt.Println(result)  // {"label": "POSITIVE", "score": 0.9998}
}
```

```python
# sentiment.py - embedded in Go binary
from transformers import pipeline
classifier = pipeline("sentiment-analysis")

def analyze(text):
    return classifier(text)[0]
```

## What This Gives You

**Single binary deployment.** Your Go app ships as one executable. On first run, it bootstraps a complete Python environment with any packages you need. Users don't install Python separately.

**Full Python ecosystem.** NumPy, PyTorch, TensorFlow, Hugging Face, OpenCV, pandas, scikit-learn—if it's on PyPI or conda-forge, you can use it.

**Multiple Python versions, simultaneously.** Run Python 3.9 for one task and Python 3.12 for another in the same application. Each environment is completely independent with its own packages.

```go
// Legacy code needs Python 3.8 with older numpy
legacyEnv, _ := jumpboot.CreateEnvironmentMamba("legacy", "./envs", "3.8", "conda-forge", nil)
legacyEnv.PipInstallPackage("numpy==1.21.0", "", "", false, nil)

// New ML pipeline needs Python 3.11 with latest packages
mlEnv, _ := jumpboot.CreateEnvironmentMamba("ml", "./envs", "3.11", "conda-forge", nil)
mlEnv.PipInstallPackages([]string{"numpy==2.0", "torch", "transformers"}, "", "", false, nil)

// Run both simultaneously
legacyRepl, _ := legacyEnv.NewREPLPythonProcess(nil, nil, nil, nil)
mlRepl, _ := mlEnv.NewREPLPythonProcess(nil, nil, nil, nil)
```

**Isolated environments.** No conflicts with system Python or other apps. Pin exact package versions per environment for reproducible builds.

**Three ways to run Python:**
- **REPL** - Interactive sessions, execute code strings, get results
- **QueueProcess** - Bidirectional RPC, call Python functions with structured data
- **Direct execution** - Run scripts, capture output

## Installation

```bash
go get github.com/richinsley/jumpboot
```

## Quick Examples

### Call Python Functions from Go

```go
env, _ := jumpboot.CreateEnvironmentMamba("myenv", "./envs", "3.11", "conda-forge", nil)
repl, _ := env.NewREPLPythonProcess(nil, nil, nil, nil)

// Execute any Python and get results
result, _ := repl.Execute("sum([1, 2, 3, 4, 5])", true)
fmt.Println(result)  // 15

// State persists between calls
repl.Execute("import numpy as np", true)
repl.Execute("data = np.random.randn(1000)", true)
result, _ = repl.Execute("np.mean(data)", true)
```

### Embed Entire Packages

```go
//go:embed mypackage/*
var myPackageFS embed.FS

func main() {
    pkg, _ := jumpboot.NewPackageFromFS("mypackage", "mypackage", "mypackage", myPackageFS)

    program := &jumpboot.PythonProgram{
        Name:     "MyApp",
        Program:  mainModule,
        Packages: []jumpboot.Package{*pkg},
    }

    process, _, _ := env.NewPythonProcessFromProgram(program, nil, nil, false)
}
```

### Bidirectional RPC

```go
// Go calls Python
result, _ := queue.Call("process_image", 30, map[string]interface{}{
    "path": "/tmp/photo.jpg",
    "resize": []int{800, 600},
})

// Python calls Go
queue.RegisterHandler("log", func(data interface{}, id string) (interface{}, error) {
    fmt.Printf("[Python] %v\n", data)
    return "ok", nil
})
```

## How It Works

1. **First run**: Downloads micromamba (~5MB), creates a conda environment with your specified Python version, installs packages via pip/conda
2. **Subsequent runs**: Environment exists, startup is fast
3. **Communication**: Go spawns Python subprocess, communicates via pipes (MessagePack serialization)
4. **Embedded code**: Python source is base64-encoded in the Go binary, loaded via custom import hooks

No CGO required for basic operation. Optional shared memory features use CGO on Unix.

## Examples

| Example | What it demonstrates |
|---------|---------------------|
| [repl](examples/repl/) | Interactive Python from Go |
| [jsonqueueserver](examples/jsonqueueserver/) | Bidirectional RPC with MessagePack |
| [chromadb](examples/chromadb/) | Vector database for embeddings |
| [gradio](examples/gradio/) | Python web UI from Go |
| [mlx](examples/mlx/) | ML inference on Apple Silicon |
| [embedded_packages](examples/embedded_packages/) | Complex package structures |

## Documentation

| Document | Description |
|----------|-------------|
| [Quick Start](docs/QUICKSTART.md) | Get running in 5 minutes |
| [Environments](docs/ENVIRONMENTS.md) | Creating and managing Python environments |
| [Programs](docs/PROGRAMS.md) | Embedding modules and packages |
| [REPL](docs/REPL.md) | Interactive Python sessions |
| [Queue Process](docs/QUEUE.md) | Bidirectional RPC |
| [Architecture](docs/ARCHITECTURE.md) | Internal design |
| [AI Agents](docs/AGENTS.md) | Guide for AI coding assistants |

## What It Doesn't Do

- **Replace Python with Go** - This is for using Python libraries from Go, not avoiding Python
- **Require CGO** - Basic features work without CGO; shared memory is optional
- **Provide Python C API bindings** - Communication is via subprocess pipes, not embedded interpreter

## Platform Support

| Platform | Status |
|----------|--------|
| macOS (amd64, arm64) | Supported |
| Linux (amd64, arm64) | Supported |
| Windows (amd64) | Supported |

## License

MIT
