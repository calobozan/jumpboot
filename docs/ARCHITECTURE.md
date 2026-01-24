# Architecture

This document describes Jumpboot's internal design and component relationships.

## Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         Go Application                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │ PythonEnvironment│  │  PythonProgram  │  │   SharedMemory  │ │
│  │                 │  │                 │  │   (optional)    │ │
│  │ - EnvPath       │  │ - Modules       │  │                 │ │
│  │ - PythonPath    │  │ - Packages      │  │ - CreateShared  │ │
│  │ - PipPath       │  │ - KVPairs       │  │ - OpenShared    │ │
│  └────────┬────────┘  └────────┬────────┘  └─────────────────┘ │
│           │                    │                                │
│           ▼                    ▼                                │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                     PythonProcess                           ││
│  │                                                             ││
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────────────┐││
│  │  │ Stdin   │  │ Stdout  │  │ Stderr  │  │ Pipes (in/out)  │││
│  │  └─────────┘  └─────────┘  └─────────┘  └─────────────────┘││
│  └──────────────────────────┬──────────────────────────────────┘│
│                             │                                   │
│  ┌──────────────────────────┼──────────────────────────────────┐│
│  │        Process Types     │                                  ││
│  │                          ▼                                  ││
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ ││
│  │  │REPLPython   │  │QueueProcess │  │ Direct Execution    │ ││
│  │  │Process      │  │             │  │                     │ ││
│  │  │             │  │ - Call()    │  │ - RunPython()       │ ││
│  │  │ - Execute() │  │ - Handlers  │  │ - RunPythonScript() │ ││
│  │  └─────────────┘  └─────────────┘  └─────────────────────┘ ││
│  └─────────────────────────────────────────────────────────────┘│
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ Pipes
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Python Subprocess                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                   Bootstrap Process                         ││
│  │                                                             ││
│  │  1. Primary bootstrap (minimal, opens FDs)                  ││
│  │  2. Secondary bootstrap (loads modules, sets up pipes)      ││
│  │  3. Custom import system (CustomFinder, CustomLoader)       ││
│  │  4. Execute main module                                     ││
│  └─────────────────────────────────────────────────────────────┘│
│                                                                 │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │ jumpboot module │  │ Embedded code   │  │ Site packages   │ │
│  │                 │  │                 │  │                 │ │
│  │ - Pipe_in      │  │ - User modules  │  │ - numpy         │ │
│  │ - Pipe_out     │  │ - User packages │  │ - requests      │ │
│  │ - Status_in    │  │                 │  │ - etc.          │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Components

### BaseEnvironment

Container-agnostic base for conda-managed environments:

```go
type BaseEnvironment struct {
    EnvironmentName   string   // Environment identifier
    RootDir           string   // Root for all environments
    EnvPath           string   // This environment's path
    EnvBinPath        string   // bin/ or Scripts/
    EnvLibPath        string   // lib/
    MicromambaVersion Version
    MicromambaPath    string
    IsNew             bool     // Newly created?
}
```

### PythonEnvironment

Extends BaseEnvironment with Python-specific information:

```go
type PythonEnvironment struct {
    BaseEnvironment           // Embedded base

    PythonVersion     Version // e.g., 3.11.4
    PipVersion        Version
    PythonPath        string  // Path to python executable
    PythonLibPath     string  // Path to libpython
    PipPath           string
    PythonHeadersPath string  // For extensions
    SitePackagesPath  string
}
```

### Runtime Interface

Defines common operations for any runtime:

```go
type Runtime interface {
    Name() string
    Path() string
    BinPath() string
    Freeze(filePath string) error
}
```

### PythonProcess

Base process type with pipes and I/O:

```go
type PythonProcess struct {
    Cmd        *exec.Cmd
    Stdin      io.WriteCloser
    Stdout     io.ReadCloser
    Stderr     io.ReadCloser
    PipeIn     io.ReadCloser   // Python → Go
    PipeOut    io.WriteCloser  // Go → Python
    StatusChan chan string
    ExceptionChan chan PythonException
}
```

### REPLPythonProcess

Interactive code execution:

```go
type REPLPythonProcess struct {
    *PythonProcess
    mutex sync.Mutex  // Serializes Execute calls
}

func (r *REPLPythonProcess) Execute(code string, waitForResult bool) (string, error)
```

### QueueProcess

Bidirectional RPC:

```go
type QueueProcess struct {
    *PythonProcess
    serializer      Serializer
    transport       Transport
    responseMap     map[string]chan map[string]interface{}
    commandHandlers map[string]CommandHandler
    methodCache     map[string]MethodInfo
}

func (q *QueueProcess) Call(method string, timeout int, args interface{}) (interface{}, error)
```

## Communication Pipes

```
Go Process                          Python Process
    │                                     │
    │──── pipeout (Go→Python) ───────────▶│
    │                                     │
    │◀─── pipein (Python→Go) ────────────│
    │                                     │
    │◀─── status (exceptions/exit) ───────│
    │                                     │
    │──── bootstrap (one-time) ──────────▶│
    │                                     │
    │──── program (one-time) ─────────────▶│
```

## Bootstrap Sequence

1. Go creates 5 pipe pairs
2. Go spawns Python with primary bootstrap script
3. Primary bootstrap opens FDs, executes secondary bootstrap
4. Secondary bootstrap:
   - Reads program JSON from pipe
   - Sets up CustomFinder/CustomLoader
   - Loads embedded modules
   - Configures jumpboot module with pipes
   - Starts watchdog thread
   - Executes main module

## Serialization

### Transport Interface

```go
type Transport interface {
    Send(data []byte) error
    Receive() ([]byte, error)
    Flush() error
}
```

### Serializer Interface

```go
type Serializer interface {
    Marshal(v interface{}) ([]byte, error)
    Unmarshal(data []byte, v interface{}) error
}
```

Default: MessagePack with length-prefixed framing.

## Thread Safety

| Component | Thread-Safe | Notes |
|-----------|-------------|-------|
| BufferPool | Yes | Channel-based |
| QueueProcess | Yes | Mutex-protected, concurrent handlers |
| REPLPythonProcess | Yes | Mutex-protected Execute() |
| SharedMemory | No | External synchronization required |
| PythonEnvironment | Yes | Read-only after creation |

## Module Loading

Embedded modules use a custom import system:

1. **CustomFinder**: Prepended to `sys.meta_path`, intercepts imports
2. **CustomLoader**: Decodes base64 source, compiles, executes
3. **linecache**: Populated for proper tracebacks

```python
class CustomFinder:
    def find_spec(self, fullname, path, target=None):
        if fullname in modules:
            return ModuleSpec(fullname, CustomLoader())
        return None

class CustomLoader:
    def exec_module(self, module):
        source = base64.b64decode(modules[module.__name__]["source"])
        exec(compile(source, module.__file__, "exec"), module.__dict__)
```

## Error Handling

Python exceptions are captured and sent via status pipe:

```go
type PythonException struct {
    Exception     string           // Exception type
    Message       string           // Exception message
    Traceback     string           // Full traceback
    Cause         *PythonException // Chained exception (from X)
    ExceptionArgs []interface{}    // Exception constructor args
}
```

Go receives exceptions via `ExceptionChan` channel.

## Environment Creation Flow

```
CreateEnvironmentMamba()
    │
    ├─► ensureMicromamba()
    │       - Download if missing
    │       - Verify version
    │
    ├─► createCondaEnvironment()
    │       - micromamba create
    │       - Install Python
    │
    └─► inspectPythonRuntime()
            - Query Python version
            - Find pip, site-packages
            - Return PythonEnvironment
```

## File Structure

```
jumpboot/
├── environment.go      # Environment types and creation
├── pyproc.go          # PythonProcess, bootstrap
├── pyprocrepl.go      # REPLPythonProcess
├── pyprocqueue.go     # QueueProcess
├── pyprocexec.go      # Direct execution helpers
├── pip.go             # Package installation
├── micromamba.go      # Micromamba management
├── shmi.go            # SharedMemory interface
├── shmi_*.go          # Platform-specific shared memory
├── ipcinterfaces.go   # Transport, Serializer
├── pyexception.go     # Exception handling
├── bufferpool.go      # Buffer pool for pipes
├── scripts/
│   ├── bootstrap.py           # Primary bootstrap
│   └── secondaryBootstrapScript.py  # Secondary bootstrap
└── jumpboot/          # Embedded Python package
    ├── __init__.py
    ├── queueserver.py
    └── sharedmemory.py
```
