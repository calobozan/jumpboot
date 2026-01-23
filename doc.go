// Package jumpboot provides seamless Python environment management and Go-Python
// interoperability without requiring CGO for core operations.
//
// Jumpboot enables Go applications to create isolated Python environments, run Python
// code, and communicate bidirectionally with Python processes using pipes and optional
// shared memory. It supports Windows, macOS, and Linux platforms.
//
// # Architecture Overview
//
// Jumpboot uses a two-stage bootstrap mechanism to launch Python processes:
//
//  1. Primary Bootstrap: A minimal Python script that opens file descriptors and
//     executes the secondary bootstrap via exec().
//
//  2. Secondary Bootstrap: Initializes the custom module import system, loads
//     embedded packages/modules, and executes the main program.
//
// This design allows Go applications to embed Python code directly in the binary
// using go:embed and have it executed transparently by the Python subprocess.
//
// # Environment Management
//
// Jumpboot provides multiple ways to create Python environments:
//
//	// Create a new environment using micromamba (auto-downloads if needed)
//	env, err := jumpboot.CreateEnvironmentMamba("myenv", "/path/to/root", "3.10", "conda-forge", nil)
//
//	// Use the system Python installation
//	env, err := jumpboot.CreateEnvironmentFromSystem()
//
//	// Create a virtual environment from an existing Python
//	env, err := jumpboot.CreateVenvEnvironment(baseEnv, "/path/to/venv", jumpboot.VenvOptions{}, nil)
//
//	// Restore an environment from a frozen JSON specification
//	env, err := jumpboot.CreateEnvironmentFromJSONFile("env.json", "/path/to/root", nil)
//
// # Process Types
//
// Jumpboot offers three process types for different interaction patterns:
//
// REPLPythonProcess provides stateful, interactive Python execution where state
// persists between calls:
//
//	repl, err := env.NewREPLPythonProcess(nil, nil, nil, nil)
//	result, err := repl.Execute("x = 42", true)
//	result, err := repl.Execute("print(x * 2)", true)  // prints 84
//	repl.Close()
//
// QueueProcess provides bidirectional RPC-style communication using MessagePack:
//
//	queue, err := env.NewQueueProcess(program, nil, nil, nil)
//	result, err := queue.Call("my_python_method", 30, map[string]interface{}{"arg": "value"})
//	queue.Close()
//
// PythonExecProcess provides simple command execution with JSON communication:
//
//	exec, err := env.NewPythonExecProcess(nil, nil)
//	output, err := exec.Exec("print('hello')")
//	exec.Close()
//
// # Embedded Modules and Packages
//
// Python code can be embedded in Go binaries using go:embed:
//
//	//go:embed scripts/mymodule.py
//	var myModuleSource string
//
//	module := jumpboot.NewModuleFromString("mymodule", "scripts/mymodule.py", myModuleSource)
//
//	//go:embed packages/mypackage/*
//	var myPackageFS embed.FS
//
//	pkg, err := jumpboot.NewPackageFromFS("mypackage", "mypackage", "packages/mypackage", myPackageFS)
//
// These modules become importable in Python via a custom import system that intercepts
// imports before the standard library.
//
// # Shared Memory (Optional, requires CGO)
//
// For high-performance data sharing, jumpboot provides cross-platform shared memory:
//
//	// Create shared memory from Go
//	shm, err := jumpboot.CreateSharedMemory("my_shm", 1024*1024)
//	defer shm.Close()
//
//	// Write data
//	shm.Write([]byte("hello from Go"))
//
//	// Get typed slices for zero-copy access
//	floats := shm.GetFloat32Slice(0)
//
// Python can access the same memory using the jumpboot.SharedMemory class.
//
// # Semaphores (Optional, requires CGO)
//
// Named semaphores enable cross-process synchronization:
//
//	sem, err := jumpboot.CreateSemaphore("my_sem", 1)
//	defer sem.Close()
//
//	sem.Acquire()
//	// critical section
//	sem.Release()
//
// # Package Installation
//
// Environments support package installation via pip and conda/micromamba:
//
//	// Install pip packages
//	err := env.PipInstallPackages([]string{"numpy", "pandas"}, "", "", true, nil)
//
//	// Install from requirements file
//	err := env.PipInstallRequirements("requirements.txt", nil)
//
//	// Install conda packages (micromamba environments only)
//	err := env.MicromambaInstallPackage("scipy", "conda-forge")
//
// # Environment Freezing
//
// Environments can be frozen to JSON for reproducibility:
//
//	err := env.FreezeToFile("environment.json")
//
// The JSON specification includes conda packages, pip packages, channels, and
// Python version, allowing exact environment recreation.
//
// # Key-Value Data Passing
//
// Data can be passed from Go to Python at process startup via KVPairs:
//
//	program := &jumpboot.PythonProgram{
//	    // ...
//	    KVPairs: map[string]interface{}{
//	        "config_path": "/path/to/config",
//	        "debug_mode":  true,
//	    },
//	}
//
// In Python, these are accessible as attributes on the jumpboot module:
//
//	import jumpboot
//	print(jumpboot.config_path)  # /path/to/config
//	print(jumpboot.debug_mode)   # True
//
// # Debugging Support
//
// Python processes can be started with debugpy for remote debugging:
//
//	program := &jumpboot.PythonProgram{
//	    // ...
//	    DebugPort:    5678,
//	    BreakOnStart: true,
//	}
//
// The process will wait for a debugger to attach before executing the main program.
//
// # Platform Support
//
// Jumpboot supports:
//   - Linux (amd64, arm64)
//   - macOS (amd64, arm64/Apple Silicon)
//   - Windows (amd64)
//
// Shared memory and semaphores use platform-specific implementations:
//   - Linux/macOS: POSIX shared memory (shm_open/mmap) and semaphores
//   - Windows: Named file mappings and kernel synchronization objects
package jumpboot
