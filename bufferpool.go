package jumpboot

// BufferPool manages a pool of byte slices to reduce allocations
type BufferPool struct {
	pool    chan []byte
	bufSize int
}

// NewBufferPool creates a new buffer pool with the given buffer size and count
func NewBufferPool(bufSize, count int) *BufferPool {
	pool := make(chan []byte, count)
	for i := 0; i < count; i++ {
		pool <- make([]byte, bufSize)
	}
	return &BufferPool{
		pool:    pool,
		bufSize: bufSize,
	}
}

// Get returns a buffer from the pool or creates a new one if none are available
func (bp *BufferPool) Get() []byte {
	select {
	case buf := <-bp.pool:
		return buf
	default:
		// If no buffers are available, allocate a new one
		return make([]byte, bp.bufSize)
	}
}

// Put returns a buffer to the pool if it's the right size
func (bp *BufferPool) Put(buf []byte) {
	if cap(buf) != bp.bufSize {
		return // Wrong size, don't put it back
	}

	// Only put it back if there's room in the pool
	select {
	case bp.pool <- buf[:bp.bufSize]: // Reset slice length and put it back
	default:
		// Pool is full, just let it be garbage collected
	}
}
