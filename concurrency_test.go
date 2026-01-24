package jumpboot

import (
	"sync"
	"testing"
)

// TestBufferPoolConcurrent tests that BufferPool is safe for concurrent access.
func TestBufferPoolConcurrent(t *testing.T) {
	pool := NewBufferPool(1024, 10)

	var wg sync.WaitGroup
	numGoroutines := 100
	numOps := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				buf := pool.Get()
				if len(buf) != 1024 {
					t.Errorf("Expected buffer length 1024, got %d", len(buf))
				}
				// Simulate some work
				buf[0] = byte(j)
				pool.Put(buf)
			}
		}()
	}

	wg.Wait()
}

// TestBufferPoolGetPutConcurrent tests concurrent Get and Put operations.
func TestBufferPoolGetPutConcurrent(t *testing.T) {
	pool := NewBufferPool(512, 5)

	var wg sync.WaitGroup
	buffers := make(chan []byte, 1000)

	// Start goroutines that Get buffers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				buf := pool.Get()
				buffers <- buf
			}
		}()
	}

	// Start goroutines that Put buffers back
	wg.Add(1)
	go func() {
		defer wg.Done()
		count := 0
		for buf := range buffers {
			pool.Put(buf)
			count++
			if count >= 500 {
				close(buffers)
				return
			}
		}
	}()

	wg.Wait()
}

// TestBufferPoolWrongSizeBuffer tests that buffers with wrong capacity are discarded.
func TestBufferPoolWrongSizeBuffer(t *testing.T) {
	pool := NewBufferPool(1024, 2)

	// Get the original buffers
	buf1 := pool.Get()
	buf2 := pool.Get()

	// Put them back
	pool.Put(buf1)
	pool.Put(buf2)

	// Try to put a wrong-sized buffer
	wrongBuf := make([]byte, 512)
	pool.Put(wrongBuf)

	// Should still be able to get two buffers
	_ = pool.Get()
	_ = pool.Get()

	// Third get should allocate new buffer (pool is empty)
	buf3 := pool.Get()
	if cap(buf3) != 1024 {
		t.Errorf("Expected new buffer with capacity 1024, got %d", cap(buf3))
	}
}
