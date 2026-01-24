# AGENTS.md - Codebase Guide for AI Assistants

This document provides guidance for AI coding assistants (Claude, Gemini, GPT, Copilot, etc.) working with the jumpboot codebase.

## Project Overview

**jumpboot** enables Go programs to run embedded Python code with full environment management. It provides:
- Automatic micromamba/conda environment creation
- Python subprocess management with IPC
- Embedded Python packages via `go:embed`
- Bidirectional RPC between Go and Python

## Architecture

### Layer Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                         │
│         (User code using jumpboot API)                       │
├─────────────────────────────────────────────────────────────┤
│                    Process Layer                             │
│   PythonProcess │ QueueProcess │ REPLPythonProcess          │
├─────────────────────────────────────────────────────────────┤
│                    Environment Layer                         │
│      PythonEnvironment │ BaseEnvironment │ Runtime          │
├─────────────────────────────────────────────────────────────┤
│                    Transport Layer                           │
│         Serializer │ Transport │ BufferPool                 │
└─────────────────────────────────────────────────────────────┘
```

### Key Types

| Type | File | Purpose |
|------|------|---------|
| `PythonEnvironment` | environment.go | Python environment with paths, versions |
| `BaseEnvironment` | environment.go | Container-agnostic conda environment base |
| `Runtime` | environment.go | Interface for multi-runtime support |
| `PythonProcess` | pyproc.go | Running Python subprocess with pipes |
| `QueueProcess` | pyprocqueue.go | Bidirectional RPC communication |
| `REPLPythonProcess` | pyprocrepl.go | Interactive REPL process |
| `PythonException` | pyexception.go | Python exception with traceback |

## Key Files

| File | Description |
|------|-------------|
| `environment.go` | Environment creation, freeze/restore, Runtime interface |
| `pyproc.go` | Core PythonProcess and bootstrap mechanism |
| `pyprocqueue.go` | Bidirectional RPC via QueueProcess |
| `pyprocrepl.go` | Interactive REPL process |
| `pip.go` | Package installation via pip |
| `micromamba.go` | Micromamba download and package installation |
| `pyexception.go` | Python exception handling |
| `ipcinterfaces.go` | Transport/Serializer abstractions |
| `bufferpool.go` | Reusable buffer pool for IPC |

## Build & Test

```bash
# Build all packages
go build ./...

# Run all tests
go test ./...

# Run specific test
go test -v -run TestCreateEnvironmentMamba

# Build examples
go build ./examples/...
```

## Common Operations

### Create a Python Environment

```go
env, err := jumpboot.CreateEnvironmentMamba(
    "myenv",           // environment name
    "./envs",          // root directory
    "3.10",            // Python version
    "conda-forge",     // conda channel
    nil,               // progress callback
)
```

### Run Python Code (REPL)

```go
repl, err := env.NewREPLPythonProcess(nil, nil, nil, nil)
result, err := repl.Execute("2 + 2", true)
repl.Close()
```

### Bidirectional RPC (QueueProcess)

```go
queue, err := env.NewQueueProcess(program, nil, nil, nil)
result, err := queue.Call("python_method", 30, args)
queue.Close()
```

### Install Packages

```go
// pip packages
err := env.PipInstallPackage("requests", "", "", false, nil)

// conda packages
err := env.MicromambaInstallPackage("numpy", "conda-forge")
```

## Thread Safety

- `BufferPool`: Thread-safe (channel-based, lock-free)
- `QueueProcess`: Thread-safe (mutex-protected, concurrent handlers)
- `REPLPythonProcess`: Thread-safe (mutex-protected, serialized execution)

## Bootstrap Mechanism

Python processes use a two-stage bootstrap:
1. **Primary bootstrap** (`scripts/bootstrap.py`): Sets up file descriptors
2. **Secondary bootstrap** (`scripts/secondaryBootstrapScript.py`): Custom import system

Embedded packages are base64-encoded and transmitted via pipes.

## Adding New Features

When modifying the codebase:
1. Update method receivers if adding to `PythonEnvironment`
2. Add tests in corresponding `*_test.go` files
3. Update examples if API changes
4. Run `go build ./...` and `go test ./...`

## Code Style

- All exported types and functions must have GoDoc comments
- Use `interface{}` (or `any` in Go 1.18+) for generic data
- Prefer channel-based concurrency over mutexes where possible
- Handle Windows path differences in platform-specific files
