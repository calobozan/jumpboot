import threading

class BufferPool:
    def __init__(self, buffer_size=8192, pool_size=10):
        """
        Initialize a pool of reusable buffers.
        
        Args:
            buffer_size: Size of each buffer in bytes
            pool_size: Number of buffers to pre-allocate
        """
        self.buffer_size = buffer_size
        self.buffers = [bytearray(buffer_size) for _ in range(pool_size)]
        self.available = []
        
        # Initially all buffers are available
        for buf in self.buffers:
            self.available.append(buf)
        
        # Lock for thread safety
        self._lock = threading.Lock()
    
    def get(self):
        """
        Get a buffer from the pool.
        
        Returns:
            A pre-allocated buffer or a new one if none are available
        """
        with self._lock:
            if self.available:
                return self.available.pop()
            else:
                # Create a new buffer if none are available
                return bytearray(self.buffer_size)
    
    def release(self, buffer):
        """
        Return a buffer to the pool.
        
        Args:
            buffer: The buffer to return to the pool
        """
        with self._lock:
            # Only add back buffers of the right size
            if len(buffer) == self.buffer_size:
                self.available.append(buffer)