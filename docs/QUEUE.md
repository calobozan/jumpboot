# Queue Process

QueueProcess provides bidirectional RPC communication between Go and Python using MessagePack serialization.

## Overview

Unlike REPL (which sends code strings), QueueProcess calls named methods with structured arguments and receives structured responses. This is useful for:

- Calling Python functions with complex arguments
- Receiving structured data from Python
- Bidirectional communication (Python can call Go)
- Long-running services with multiple operations

## Basic Usage

### Python Side

```python
# server.py
from jumpboot.queueserver import MessagePackQueueServer

server = MessagePackQueueServer()

@server.register
def add(a: int, b: int) -> int:
    """Add two numbers."""
    return a + b

@server.register
def process_data(items: list) -> dict:
    """Process a list and return statistics."""
    return {
        "count": len(items),
        "sum": sum(items),
        "mean": sum(items) / len(items) if items else 0
    }

if __name__ == "__main__":
    server.run()
```

### Go Side

```go
package main

import (
    _ "embed"
    "fmt"
    "log"

    "github.com/richinsley/jumpboot"
)

//go:embed server.py
var serverScript string

func main() {
    env, _ := jumpboot.CreateEnvironmentMamba("queue-env", "./envs", "3.11", "conda-forge", nil)

    // Create program with the server script
    program := jumpboot.CreateProgramFromString("server", "server.py", serverScript, nil, nil)

    // Start queue process
    queue, _ := env.NewQueueProcess(program, nil, nil, nil)
    defer queue.Close()

    // Call Python methods
    result, _ := queue.Call("add", 10, map[string]interface{}{"a": 5, "b": 3})
    fmt.Println(result) // 8

    result, _ = queue.Call("process_data", 10, map[string]interface{}{
        "items": []int{1, 2, 3, 4, 5},
    })
    fmt.Printf("%v\n", result) // map[count:5 sum:15 mean:3]
}
```

## Fluent API

QueueProcess provides a fluent builder for method calls:

```go
// Simple call
result, err := queue.On("add").Do("a", 5, "b", 3).Call()

// With timeout (seconds)
result, err := queue.On("slow_operation").
    Do("input", data).
    WithTimeout(30 * time.Second).
    Call()

// Unmarshal into struct
var stats Statistics
err := queue.On("process_data").
    Do("items", items).
    CallReflect(&stats)
```

## Bidirectional Communication

Python can call registered Go handlers:

```go
// Register a Go handler
queue.RegisterHandler("log_message", func(data interface{}, requestID string) (interface{}, error) {
    msg := data.(map[string]interface{})
    fmt.Printf("[Python] %s: %s\n", msg["level"], msg["message"])
    return "logged", nil
})
```

```python
# In Python
@server.register
def do_work():
    # Call back to Go
    server.call_go("log_message", {"level": "INFO", "message": "Starting work"})
    result = expensive_operation()
    server.call_go("log_message", {"level": "INFO", "message": "Done"})
    return result
```

## Service Struct Pattern

Register all methods of a Go struct as handlers:

```go
type MyService struct{}

func (s *MyService) FetchConfig(key string) (string, error) {
    return os.Getenv(key), nil
}

func (s *MyService) SaveResult(data map[string]interface{}) (bool, error) {
    // Save to database
    return true, nil
}

// All exported methods become handlers
queue, _ := env.NewQueueProcess(program, &MyService{}, nil, nil)
```

Python can then call `FetchConfig` and `SaveResult` directly.

## Method Discovery

QueueProcess discovers Python methods on startup:

```go
// List available methods
methods := queue.GetMethods()
for _, name := range methods {
    fmt.Println(name)
}

// Get method info (parameters, docs)
info, ok := queue.GetMethodInfo("process_data")
if ok {
    fmt.Printf("Doc: %s\n", info.Doc)
    for _, param := range info.Parameters {
        fmt.Printf("  %s (%s) required=%v\n", param.Name, param.Type, param.Required)
    }
}
```

## Shutdown

```go
// Graceful shutdown (waits for Python cleanup)
queue.Shutdown()

// Immediate termination
queue.Close()
```

## Thread Safety

QueueProcess is safe for concurrent use:

- Multiple goroutines can call `Call()` simultaneously
- Requests are serialized via mutex
- Responses are correlated with requests using unique IDs
- Command handlers run in separate goroutines

```go
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(n int) {
        defer wg.Done()
        result, _ := queue.Call("compute", 10, map[string]interface{}{"n": n})
        fmt.Printf("Result %d: %v\n", n, result)
    }(i)
}
wg.Wait()
```

## Protocol

Messages are length-prefixed MessagePack:

```
[4-byte length (big-endian)][msgpack data]
```

Request format:
```json
{
    "command": "method_name",
    "data": {"arg1": "value1"},
    "request_id": "req-1"
}
```

Response format:
```json
{
    "result": "return value",
    "request_id": "req-1"
}
```

Error format:
```json
{
    "error": "error message",
    "request_id": "req-1"
}
```
