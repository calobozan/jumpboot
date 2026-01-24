package jumpboot

// BufferPool manages a pool of reusable byte slices to reduce GC pressure.
// It uses a channel-based design for thread-safe access without locks.
//
// BufferPool is safe for concurrent use by multiple goroutines. The channel-based
// implementation provides lock-free synchronization for both Get and Put operations.
type BufferPool struct {
	pool    chan []byte
	bufSize int
}

// NewBufferPool creates a pool pre-populated with count buffers of bufSize bytes.
// Buffers are retrieved with Get and returned with Put.
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

// Get returns a buffer from the pool, or allocates a new one if the pool is empty.
// The returned buffer has capacity bufSize but length may vary.
func (bp *BufferPool) Get() []byte {
	select {
	case buf := <-bp.pool:
		return buf
	default:
		// If no buffers are available, allocate a new one
		return make([]byte, bp.bufSize)
	}
}

// Put returns a buffer to the pool for reuse.
// Buffers with incorrect capacity are discarded (not returned to pool).
// If the pool is full, the buffer is discarded for garbage collection.
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
