package jumpboot

// Serializer defines the interface for message encoding and decoding.
// Implementations convert between Go values and byte slices for transport.
// The default implementation uses MessagePack for efficient binary serialization.
type Serializer interface {
	// Marshal encodes a Go value to bytes.
	Marshal(v interface{}) ([]byte, error)

	// Unmarshal decodes bytes into a Go value.
	Unmarshal(data []byte, v interface{}) error
}

// Transport defines the interface for sending and receiving byte messages.
// Implementations handle the wire protocol (framing, buffering, etc.).
// The default implementation uses length-prefixed binary messages over pipes.
type Transport interface {
	// Send transmits a message to the remote endpoint.
	Send(data []byte) error

	// Receive reads a complete message from the remote endpoint.
	Receive() ([]byte, error)

	// Close releases transport resources and closes underlying connections.
	Close() error

	// Flush ensures any buffered data is sent immediately.
	Flush() error
}
