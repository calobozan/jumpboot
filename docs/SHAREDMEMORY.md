# Shared Memory

SharedMemory provides zero-copy data transfer between Go and Python for high-performance scenarios.

## Requirements

- **Linux/macOS**: Uses POSIX shared memory (shm_open, mmap). Requires CGO.
- **Windows**: Uses named file mappings. No CGO required.

Build with CGO enabled:
```bash
CGO_ENABLED=1 go build
```

## Basic Usage

### Go Side (Creator)

```go
package main

import (
    "fmt"
    "github.com/richinsley/jumpboot"
)

func main() {
    // Create 1MB shared memory region
    shm, err := jumpboot.CreateSharedMemory("/my_data", 1024*1024)
    if err != nil {
        panic(err)
    }
    defer shm.Close()

    // Write data
    data := []byte("Hello from Go!")
    shm.Write(data)

    // Reset position and read back
    shm.Seek(0, 0)
    buf := make([]byte, len(data))
    shm.Read(buf)
    fmt.Println(string(buf))
}
```

### Python Side (Consumer)

```python
from jumpboot import SharedMemory

# Open existing shared memory
shm = SharedMemory("/my_data", 1024*1024)

# Read data
data = shm.read(14)
print(data)  # b"Hello from Go!"

# Write response
shm.seek(0)
shm.write(b"Hello from Python!")

shm.close()
```

## Typed Slice Access

Get zero-copy typed slices for direct memory access:

```go
shm, _ := jumpboot.CreateSharedMemory("/floats", 1024*8)

// Get float64 slice (zero-copy view of memory)
floats := shm.GetFloat64Slice(0)
floats[0] = 3.14159
floats[1] = 2.71828

// Changes are immediately visible in shared memory
```

Available typed slices:
- `GetFloat32Slice(offset)` / `GetFloat64Slice(offset)`
- `GetInt16Slice(offset)` / `GetInt32Slice(offset)` / `GetInt64Slice(offset)`
- `GetUint16Slice(offset)` / `GetUint32Slice(offset)` / `GetUint64Slice(offset)`
- `GetByteSlice(offset)`

Generic version:
```go
// Any type
slice := jumpboot.GetTypedSlice[float32](shm, offset)
```

## NumPy Integration

Share NumPy arrays between Go and Python with zero copy:

### Go Side

```go
// Create shared memory sized for a 1000x1000 float64 array
size := 1000 * 1000 * 8
shm, _ := jumpboot.CreateSharedMemory("/numpy_array", size)

// Get as float64 slice and populate
data := shm.GetFloat64Slice(0)
for i := range data[:1000*1000] {
    data[i] = float64(i)
}
```

### Python Side

```python
import numpy as np
from jumpboot import SharedMemory

shm = SharedMemory("/numpy_array", 1000*1000*8)

# Create NumPy array backed by shared memory
arr = np.frombuffer(shm.buffer, dtype=np.float64).reshape(1000, 1000)

# Operations on arr modify shared memory directly
arr *= 2.0  # Go can now see the doubled values

# Compute and store result
result = np.sum(arr)
```

## io.Reader/Writer Interface

SharedMemory implements standard Go interfaces:

```go
// io.Reader
n, err := shm.Read(buf)

// io.Writer
n, err := shm.Write(data)

// io.Seeker
pos, err := shm.Seek(offset, io.SeekStart)

// io.ReaderAt
n, err := shm.ReadAt(buf, offset)

// io.WriterAt
n, err := shm.WriteAt(data, offset)
```

## Synchronization

Shared memory requires explicit synchronization between processes. Common patterns:

### Using QueueProcess for Coordination

```go
// Go: write data, then signal Python
shm.Write(largeData)
queue.Call("process_buffer", 10, map[string]interface{}{
    "shm_name": "/my_data",
    "size": len(largeData),
})
```

```python
@server.register
def process_buffer(shm_name: str, size: int):
    shm = SharedMemory(shm_name, size)
    data = shm.read(size)
    # Process data...
    shm.close()
    return {"status": "done"}
```

### Using Files or Signals

For producer-consumer patterns, use a separate coordination mechanism (pipes, files, or the queue process).

## Memory Layout

When sharing structured data, agree on the layout:

```go
// Header: 8 bytes for size, then data
binary.BigEndian.PutUint64(shm.GetByteSlice(0), uint64(dataLen))
shm.WriteAt(data, 8)
```

```python
import struct
size = struct.unpack('>Q', shm.read(8))[0]
data = shm.read(size)
```

## Naming Conventions

- On POSIX systems, names should start with `/`
- Names are global to the system
- Use unique names to avoid conflicts

```go
name := fmt.Sprintf("/jumpboot_%d_%s", os.Getpid(), uuid.New().String()[:8])
shm, _ := jumpboot.CreateSharedMemory(name, size)
```

## Cleanup

Shared memory persists until explicitly destroyed or system reboot:

```go
shm.Close()  // Unmaps from this process
```

On POSIX, the memory segment is destroyed when all processes close it. On Windows, named mappings are reference-counted.

## Limitations

- Requires CGO on Unix (Linux/macOS)
- Size must be agreed upon by both processes
- No built-in synchronization (use external coordination)
- Memory is not automatically initialized to zero on all platforms
