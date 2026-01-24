# Quick Start

This guide gets you running Python from Go in under 5 minutes.

## Install

```bash
go get github.com/richinsley/jumpboot
```

## Minimal Example

```go
package main

import (
    "fmt"
    "log"

    "github.com/richinsley/jumpboot"
)

func main() {
    // Create a Python 3.11 environment using micromamba
    env, err := jumpboot.CreateEnvironmentMamba("myenv", "./envs", "3.11", "conda-forge", nil)
    if err != nil {
        log.Fatal(err)
    }

    // Start an interactive Python session
    repl, _ := env.NewREPLPythonProcess(nil, nil, nil, nil)
    defer repl.Close()

    // Execute Python code
    result, _ := repl.Execute("2 + 2", true)
    fmt.Println(result) // 4

    result, _ = repl.Execute("import sys; sys.version", true)
    fmt.Println(result) // '3.11.x ...'
}
```

## What Happens

1. **First run**: Jumpboot downloads micromamba (~5MB) and creates a conda environment with Python 3.11. This takes a minute or two.

2. **Subsequent runs**: The environment already exists, so startup is fast.

3. **REPL session**: A Python subprocess starts with pipes for communication. `Execute()` sends code and returns the result.

## Environment Location

Environments are created under the path you specify:

```
./envs/
  bin/
    micromamba
  envs/
    myenv/
      bin/
        python
      lib/
        python3.11/
          site-packages/
```

## Installing Packages

```go
// On first run
if env.IsNew {
    env.PipInstallPackage("requests", "", "", false, nil)
    env.PipInstallPackage("numpy", "", "", false, nil)
}
```

Or install multiple packages:

```go
env.PipInstallPackages([]string{"requests", "numpy", "pandas"}, "", "", false, nil)
```

## Embedding Python Code

Use `go:embed` to include Python code in your binary:

```go
//go:embed myscript.py
var myScript string

func main() {
    // ...
    mod := jumpboot.NewModuleFromString("myscript", "myscript.py", myScript)
    repl, _ := env.NewREPLPythonProcess(nil, nil, []jumpboot.Module{*mod}, nil)

    repl.Execute("import myscript", true)
    result, _ := repl.Execute("myscript.my_function()", true)
}
```

## Next Steps

- [Environments](ENVIRONMENTS.md) - Different environment types and management
- [Programs](PROGRAMS.md) - Embedding complex packages
- [REPL](REPL.md) - Interactive session details
- [Queue Process](QUEUE.md) - Bidirectional RPC
