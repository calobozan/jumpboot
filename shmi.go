package jumpboot

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"
)

// SharedMemory provides cross-platform shared memory for efficient data exchange
// between Go and Python processes. It implements io.Reader, io.Writer, io.Seeker,
// io.ReaderAt, and io.WriterAt for flexible access patterns.
//
// Shared memory is created with CreateSharedMemory and opened with OpenSharedMemory.
// Both processes must use the same name and agree on the size.
//
// Note: This feature requires CGO and platform-specific implementations:
//   - Linux/macOS: POSIX shared memory (shm_open, mmap)
//   - Windows: Named file mappings
//
// Example:
//
//	// In Go (creator)
//	shm, _ := jumpboot.CreateSharedMemory("/my_shm", 1024*1024)
//	shm.Write([]byte("hello"))
//	shm.Close()
//
//	// In Python (consumer)
//	from jumpboot import SharedMemory
//	shm = SharedMemory("/my_shm", 1024*1024)
//	data = shm.read(5)  # b"hello"
type SharedMemory struct {
	// m is the platform-specific shared memory implementation
	m *shmi

	// pos is the current read/write position
	pos int64

	// Name is the identifier used to open/create this shared memory
	Name string
}

// GetSize returns the size of the shared memory region in bytes.
func (o *SharedMemory) GetSize() int {
	return o.m.getSize()
}

// GetPtr returns an unsafe pointer to the shared memory region.
// Use with caution; prefer the typed slice methods for safer access.
func (o *SharedMemory) GetPtr() unsafe.Pointer {
	return o.m.getPtr()
}

// CreateSharedMemory creates a new named shared memory region.
// The name should start with "/" on POSIX systems. The size is in bytes.
// Returns an error if the memory cannot be allocated or mapped.
func CreateSharedMemory(name string, size int) (*SharedMemory, error) {
	m, err := create(name, size)
	if err != nil {
		return nil, err
	}
	return &SharedMemory{m, 0, name}, nil
}

// OpenSharedMemory opens an existing named shared memory region.
// The name and size must match those used when creating the memory.
// Returns an error if the memory doesn't exist or cannot be mapped.
func OpenSharedMemory(name string, size int) (*SharedMemory, error) {
	m, err := open(name, size)
	if err != nil {
		return nil, err
	}
	return &SharedMemory{m, 0, name}, nil
}

// Close unmaps and releases the shared memory region.
// The underlying memory is only destroyed when all processes have closed it.
func (o *SharedMemory) Close() (err error) {
	if o.m != nil {
		err = o.m.close()
		if err == nil {
			o.m = nil
		}
	}
	return err
}

// Read reads up to len(p) bytes from shared memory at the current position.
// Implements io.Reader.
func (o *SharedMemory) Read(p []byte) (n int, err error) {
	n, err = o.ReadAt(p, o.pos)
	if err != nil {
		return 0, err
	}
	o.pos += int64(n)
	return n, nil
}

// ReadAt reads len(p) bytes from shared memory starting at offset off.
// Implements io.ReaderAt.
func (o *SharedMemory) ReadAt(p []byte, off int64) (n int, err error) {
	return o.m.readAt(p, off)
}

// Seek sets the position for the next Read or Write.
// Implements io.Seeker with io.SeekStart, io.SeekCurrent, and io.SeekEnd.
func (o *SharedMemory) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		offset += int64(0)
	case io.SeekCurrent:
		offset += o.pos
	case io.SeekEnd:
		offset += int64(o.m.size)
	}
	if offset < 0 || offset >= int64(o.m.size) {
		return 0, fmt.Errorf("invalid offset")
	}
	o.pos = offset
	return offset, nil
}

// Write writes len(p) bytes to shared memory at the current position.
// Implements io.Writer.
func (o *SharedMemory) Write(p []byte) (n int, err error) {
	n, err = o.WriteAt(p, o.pos)
	if err != nil {
		return 0, err
	}
	o.pos += int64(n)
	return n, nil
}

// WriteAt writes len(p) bytes to shared memory starting at offset off.
// Implements io.WriterAt.
func (o *SharedMemory) WriteAt(p []byte, off int64) (n int, err error) {
	return o.m.writeAt(p, off)
}

// GetTypedSlice returns a typed slice view of shared memory starting at offset.
// The slice provides zero-copy access to the underlying memory.
// Changes to the slice are immediately visible in shared memory.
//
// Warning: The returned slice is only valid while the SharedMemory is open.
// Using it after Close() results in undefined behavior.
func GetTypedSlice[T any](shm *SharedMemory, offset int) []T {
	// Calculate the number of elements that can fit in the remaining space
	elementSize := int(unsafe.Sizeof(*new(T)))
	remainingSize := shm.m.size - offset
	numElements := remainingSize / elementSize

	// Create a slice using unsafe.Slice
	ptr := shm.GetPtr()
	return unsafe.Slice((*T)(unsafe.Add(ptr, uintptr(offset))), int(numElements))
}

// GetFloat32Slice returns a float32 slice view of shared memory at offset.
func (o *SharedMemory) GetFloat32Slice(offset int) []float32 {
	return GetTypedSlice[float32](o, offset)
}

// GetFloat64Slice returns a float64 slice view of shared memory at offset.
func (o *SharedMemory) GetFloat64Slice(offset int) []float64 {
	return GetTypedSlice[float64](o, offset)
}

// GetInt16Slice returns an int16 slice view of shared memory at offset.
func (o *SharedMemory) GetInt16Slice(offset int) []int16 {
	return GetTypedSlice[int16](o, offset)
}

// GetInt32Slice returns an int32 slice view of shared memory at offset.
func (o *SharedMemory) GetInt32Slice(offset int) []int32 {
	return GetTypedSlice[int32](o, offset)
}

// GetInt64Slice returns an int64 slice view of shared memory at offset.
func (o *SharedMemory) GetInt64Slice(offset int) []int64 {
	return GetTypedSlice[int64](o, offset)
}

// GetUint16Slice returns a uint16 slice view of shared memory at offset.
func (o *SharedMemory) GetUint16Slice(offset int) []uint16 {
	return GetTypedSlice[uint16](o, offset)
}

// GetUint32Slice returns a uint32 slice view of shared memory at offset.
func (o *SharedMemory) GetUint32Slice(offset int) []uint32 {
	return GetTypedSlice[uint32](o, offset)
}

// GetUint64Slice returns a uint64 slice view of shared memory at offset.
func (o *SharedMemory) GetUint64Slice(offset int) []uint64 {
	return GetTypedSlice[uint64](o, offset)
}

// GetByteSlice returns a byte slice view of shared memory at offset.
func (o *SharedMemory) GetByteSlice(offset int) []byte {
	return GetTypedSlice[byte](o, offset)
}

// memcpy in go:
// https://go.dev/play/p/MFJjHhDZatl
// https://stackoverflow.com/questions/69816793/golang-fast-alternative-to-memcpy

func copySlice2Ptr(b []byte, p uintptr, off int64, size int) int {
	h := reflect.SliceHeader{}
	h.Cap = int(size)
	h.Len = int(size)
	h.Data = p

	bb := *(*[]byte)(unsafe.Pointer(&h))
	return copy(bb[off:], b)
}

func copyPtr2Slice(p uintptr, b []byte, off int64, size int) int {
	h := reflect.SliceHeader{}
	h.Cap = int(size)
	h.Len = int(size)
	h.Data = p

	bb := *(*[]byte)(unsafe.Pointer(&h))
	return copy(b, bb[off:size])
}
