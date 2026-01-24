# Jumpboot Bootstrapping Process

This document explains the internal bootstrapping process used by Jumpboot to initialize and run Python code within a Go application. Understanding this process is helpful for advanced usage and debugging.

## Overview

Jumpboot uses a two-stage bootstrapping process to execute Python code without relying on direct CGO bindings (except for the optional shared memory features). This approach enhances portability and reduces complexity. The core idea is to launch a Python subprocess and communicate with it via pipes.

The two stages are:

1. **Primary Bootstrap (Go Side):** A minimal, embedded Python script (`scripts/bootstrap.py`) is executed as the initial command passed to the Python interpreter. This script's primary responsibility is to define a cross-platform file descriptor opening helper and execute the secondary bootstrap script.

2. **Secondary Bootstrap (Python Side):** A more extensive Python script (`scripts/secondaryBootstrapScript.py`) is read from a pipe and executed. This script handles:
    * Loading the embedded Python modules and packages (including `jumpboot` itself).
    * Setting up a custom module finder (`CustomFinder`) and loader (`CustomLoader`) to handle imports from the embedded code.
    * Setting up communication pipes (`Pipe_in`, `Pipe_out`, `Status_in`) on the `jumpboot` module.
    * Starting a watchdog thread to monitor the parent process.
    * (Optionally) starting a debugpy server for debugging.
    * Executing the main Python program.

## Pipe Architecture

Jumpboot creates multiple pipes for communication:

| Pipe | Direction | Purpose |
|------|-----------|---------|
| `pipein` | Python → Go | Data from Python to Go |
| `pipeout` | Go → Python | Data from Go to Python |
| `status` | Python → Go | Status messages and exceptions |
| `bootstrap` | Go → Python | Transmits secondary bootstrap script |
| `program` | Go → Python | Transmits JSON program data |
| `stdin/stdout/stderr` | Bidirectional | Standard I/O streams |

## Detailed Steps

### 1. Primary Bootstrap Script

The primary bootstrap script (`scripts/bootstrap.py`) is extremely minimal for fast startup:

```python
import os,sys
def o(h,m='r'):
 if sys.platform.startswith('win'):import msvcrt;return os.fdopen(msvcrt.open_osfhandle(h,os.O_RDONLY if m=='r'else os.O_WRONLY),m)
 return os.fdopen(h,m)
sys.__jbo=o;exec(o(int(sys.argv[2])).read())
```

This script:
1. Defines a cross-platform helper function `o(h, m)` that opens file handles/descriptors
2. Attaches this function to `sys.__jbo` for use by the secondary bootstrap
3. Reads and executes the secondary bootstrap script from the pipe FD passed in `sys.argv[2]`

### 2. Go-Side Initialization (`NewPythonProcessFromProgram`)

When you call `NewPythonProcessFromProgram` in your Go code:

1. **Jumpboot Package Injection:** The embedded `jumpboot` Python package is prepended to the program's package list.

2. **Pipe Creation:** Five pipe pairs are created:
   - `reader_bootstrap` / `writer_bootstrap`: For the secondary bootstrap script
   - `reader_program` / `writer_program`: For the JSON program data
   - `pipein_reader` / `pipein_writer`: For Python → Go data
   - `pipeout_reader` / `pipeout_writer`: For Go → Python data
   - `status_reader` / `status_writer`: For status/exception reporting

3. **File Descriptor Assignment:** The file descriptors are passed to Python via `ExtraFiles` (Unix) or `AdditionalInheritedHandles` (Windows).

4. **Command Construction:** The Python command is built as:
   ```
   python -u -c "<primary_bootstrap>" <extra_fd_count> <fd1> <fd2> ... <user_args>
   ```

5. **Program Data:** The `PythonProgram` struct is JSON-serialized and includes:
   - `PipeIn`, `PipeOut`, `StatusIn`: File descriptor numbers
   - `Packages`: List of embedded packages with base64-encoded source
   - `Modules`: List of standalone modules
   - `Program`: The main module to execute
   - `KVPairs`: Key-value data accessible as `jumpboot.<key>`
   - `DebugPort`, `BreakOnStart`: Optional debugging configuration

6. **Goroutines:** Two goroutines write data to pipes and close them:
   - Secondary bootstrap script → `writer_bootstrap`
   - JSON program data → `writer_program`

### 3. Secondary Bootstrap Execution

Once the secondary bootstrap script starts running:

1. **Program Data Loading:** Reads JSON program data from the program pipe and deserializes it.

2. **Pipe Setup:** Opens the communication pipes using the FDs from `program_data`:
   ```python
   f_out = sys.__jbo(fd_out, 'w')  # Go → Python
   f_in = sys.__jbo(fd_in, 'r')   # Python → Go
   f_status = sys.__jbo(fd_status, 'w')  # Status/exceptions
   ```

3. **Module Processing:** The `load_program_data` function processes the program data and creates a flat dictionary of all modules with their:
   - Full dotted name (e.g., `mypackage.subpackage.module`)
   - Virtual file path (for `__file__` and tracebacks)
   - Base64-encoded source code

4. **Custom Import System:**
   - **`CustomFinder`:** Implements `MetaPathFinder`, intercepts import requests, and checks if the module exists in the loaded modules dictionary.
   - **`CustomLoader`:** Implements `Loader`, responsible for:
     - Decoding base64 source code
     - Setting module attributes (`__file__`, `__package__`, `__path__`, `__spec__`)
     - Adding source to `linecache` for proper tracebacks
     - Compiling and executing module code

5. **`sys.meta_path` Modification:** The `CustomFinder` is prepended to `sys.meta_path`, ensuring embedded modules are found before system modules.

6. **Package Initialization:** All top-level packages are loaded (except `__main__`).

7. **Jumpboot Module Setup:** The `jumpboot` module is configured with:
   ```python
   jumpboot.Pipe_in = f_in    # Read from Go
   jumpboot.Pipe_out = f_out  # Write to Go
   jumpboot.Status_in = f_status  # Status/exceptions
   # Plus any KVPairs from program_data
   ```

8. **Watchdog Thread:** A daemon thread monitors the parent Go process:
   ```python
   def watchdog_monitor_parent():
       parent_pid = os.getppid()
       while True:
           # Check if parent is still alive
           # Exit if parent dies
           time.sleep(3)
   ```

9. **Main Module Execution:** The main module is executed with proper exception handling:
   ```python
   try:
       loader.exec_module(main_module)
   except Exception as e:
       # Send exception to status pipe
       f_status.write(json.dumps(exception_info) + "\n")
   finally:
       # Send exit status
       f_status.write(json.dumps({"type": "status", "message": "exit"}) + "\n")
   ```

## Command-Line Argument Structure

The Python process receives arguments in this order:

| Position | Content | Example |
|----------|---------|---------|
| `sys.argv[0]` | Python executable | `python` |
| `sys.argv[1]` | Extra FD count | `5` |
| `sys.argv[2]` | Bootstrap pipe FD | `6` |
| `sys.argv[3]` | Program data pipe FD | `7` |
| `sys.argv[4...]` | Extra file descriptors | `8`, `9`, ... |
| Remaining | User arguments | `--verbose`, etc. |

After processing, `sys.argv` is adjusted to contain only user arguments.

## Key Concepts

* **Pipes:** The primary communication mechanism. Unidirectional byte streams that are reliable and cross-platform.

* **File Descriptors (FDs) / Handles:** Numerical identifiers for open files and pipes. On Windows, these are converted using `msvcrt.open_osfhandle()`.

* **`sys.__jbo`:** A helper function attached to `sys` that provides cross-platform file descriptor opening.

* **`sys.meta_path`:** Python's import hook list. By prepending `CustomFinder`, Jumpboot intercepts imports for embedded modules.

* **`linecache`:** Python's source cache for tracebacks. Jumpboot adds embedded source code here so tracebacks show correct file contents.

* **Watchdog Thread:** Monitors the parent Go process and exits if it dies, preventing orphaned Python processes.

## Status and Exception Reporting

The status pipe carries JSON messages from Python to Go:

**Exception message:**
```json
{
  "type": "exception",
  "exception": "ValueError",
  "message": "invalid value",
  "traceback": "Traceback (most recent call last):..."
}
```

**Exit status:**
```json
{
  "type": "status",
  "message": "exit"
}
```

Go reads these via the `StatusChan` and `ExceptionChan` channels on `PythonProcess`.

## Debugging Support

If `DebugPort` is set in the program data:

1. `debugpy` is imported (and installed if missing)
2. A debug server starts on the specified port
3. Execution pauses until a debugger connects
4. If `BreakOnStart` is true, a breakpoint is set at the first line

## Advantages of this Approach

* **No CGO (Mostly):** Avoids complexities and platform-specific issues of CGO for general Python interaction.
* **Isolation:** Python runs in a separate process, providing strong isolation.
* **Flexibility:** Supports micromamba, conda, venv, and system Python environments.
* **Portability:** Works consistently across Windows, macOS, and Linux.
* **Debuggable:** Full debugpy support for remote debugging.

## Limitations

* **Overhead:** Launching a separate process has some overhead compared to direct CGO calls. However, for many use cases, this overhead is negligible.
* **Communication:** Requires serialization (JSON, MessagePack) for structured data. The optional shared memory feature mitigates this for high-performance scenarios.
