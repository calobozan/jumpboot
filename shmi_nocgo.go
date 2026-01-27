//go:build (darwin || linux) && !cgo

package jumpboot

import (
	"errors"
	"io"
	"unsafe"
)

// ErrSharedMemoryNotAvailable is returned when shared memory operations are
// attempted but CGO is disabled. Shared memory requires CGO on Unix platforms.
var ErrSharedMemoryNotAvailable = errors.New("shared memory requires CGO on this platform; rebuild with CGO_ENABLED=1")

// shmi is a stub implementation for when CGO is disabled.
// All operations return ErrSharedMemoryNotAvailable.
type shmi struct {
	size int
}

func (o *shmi) getSize() int {
	return o.size
}

func (o *shmi) getPtr() unsafe.Pointer {
	return nil
}

func create(name string, size int) (*shmi, error) {
	return nil, ErrSharedMemoryNotAvailable
}

func open(name string, size int) (*shmi, error) {
	return nil, ErrSharedMemoryNotAvailable
}

func (o *shmi) close() error {
	return ErrSharedMemoryNotAvailable
}

func (o *shmi) readAt(p []byte, off int64) (n int, err error) {
	return 0, io.EOF
}

func (o *shmi) writeAt(p []byte, off int64) (n int, err error) {
	return 0, io.EOF
}
