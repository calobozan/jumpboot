package jumpboot

// Semaphore provides cross-process synchronization using named semaphores.
// It enables coordination between Go and Python processes accessing shared resources.
//
// Note: This feature requires CGO and platform-specific implementations:
//   - Linux/macOS: POSIX named semaphores (sem_open)
//   - Windows: Kernel semaphore objects
//
// Create a semaphore with CreateSemaphore and open an existing one with OpenSemaphore.
// Both processes must use the same name.
//
// Example:
//
//	sem, _ := jumpboot.CreateSemaphore("/my_sem", 1)
//	defer sem.Close()
//
//	sem.Acquire()
//	// critical section - access shared resource
//	sem.Release()
type Semaphore interface {
	// Acquire blocks until the semaphore can be decremented.
	Acquire() error

	// Release increments the semaphore, potentially unblocking waiters.
	Release() error

	// TryAcquire attempts to decrement the semaphore without blocking.
	// Returns true if acquired, false if the semaphore was not available.
	TryAcquire() (bool, error)

	// AcquireTimeout attempts to acquire with a maximum wait time in milliseconds.
	// Returns true if acquired, false if the timeout elapsed.
	AcquireTimeout(timeoutMs int) (bool, error)

	// Close releases resources associated with the semaphore.
	// The semaphore is only destroyed when all processes have closed it.
	Close() error
}
