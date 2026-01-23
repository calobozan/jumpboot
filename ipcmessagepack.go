package jumpboot

import (
	"encoding/binary"
	"io"

	"github.com/vmihailenco/msgpack/v5"
)

// MsgpackSerializer implements Serializer using MessagePack encoding.
// MessagePack is a binary serialization format that is more compact and faster
// than JSON while maintaining similar semantics.
type MsgpackSerializer struct{}

// Marshal encodes a Go value to MessagePack bytes.
func (ms MsgpackSerializer) Marshal(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

// Unmarshal decodes MessagePack bytes into a Go value.
func (ms MsgpackSerializer) Unmarshal(data []byte, v interface{}) error {
	return msgpack.Unmarshal(data, v)
}

// MsgpackTransport implements Transport using length-prefixed binary framing.
// Each message is sent as a 4-byte big-endian length followed by the message bytes.
// This matches the protocol used by the Python jumpboot.msgpackqueue module.
type MsgpackTransport struct {
	reader     io.ReadCloser
	writer     io.WriteCloser
	bufferPool *BufferPool
}

// NewMsgpackTransport creates a new MsgpackTransport using the provided reader and writer.
// A buffer pool with 8KB buffers (matching Python) is created for efficient memory usage.
func NewMsgpackTransport(reader io.ReadCloser, writer io.WriteCloser) *MsgpackTransport {
	return &MsgpackTransport{reader: reader,
		writer:     writer,
		bufferPool: NewBufferPool(8192, 10), // Same size as Python side
	}
}

// Send transmits a message with a 4-byte length prefix.
// The length is encoded as big-endian uint32.
func (mt *MsgpackTransport) Send(data []byte) error {
	// Get length and convert to 4-byte array
	lengthBytes := mt.bufferPool.Get()[:4]
	binary.BigEndian.PutUint32(lengthBytes, uint32(len(data)))

	// Send length
	if _, err := mt.writer.Write(lengthBytes); err != nil {
		mt.bufferPool.Put(lengthBytes)
		return err
	}

	// Flush after length to ensure it's received immediately
	if flusher, ok := mt.writer.(interface{ Flush() error }); ok {
		if err := flusher.Flush(); err != nil {
			mt.bufferPool.Put(lengthBytes)
			return err
		}
	}

	// Put length buffer back in the pool
	mt.bufferPool.Put(lengthBytes)

	// Send data
	_, err := mt.writer.Write(data)
	if err != nil {
		return err
	}

	// Flush after data
	if flusher, ok := mt.writer.(interface{ Flush() error }); ok {
		return flusher.Flush()
	}

	return nil
}

// Receive reads a length-prefixed message from the transport.
// Small messages (<=8KB) use the buffer pool; larger messages allocate new buffers.
func (mt *MsgpackTransport) Receive() ([]byte, error) {
	// Get buffer for length
	lengthBuf := mt.bufferPool.Get()[:4]

	// Read length
	if _, err := io.ReadFull(mt.reader, lengthBuf); err != nil {
		mt.bufferPool.Put(lengthBuf)
		return nil, err
	}

	length := binary.BigEndian.Uint32(lengthBuf)
	mt.bufferPool.Put(lengthBuf)

	// For small messages, use buffer pool
	if length <= uint32(mt.bufferPool.bufSize) {
		buf := mt.bufferPool.Get()[:length]
		_, err := io.ReadFull(mt.reader, buf)
		if err != nil {
			mt.bufferPool.Put(buf)
			return nil, err
		}

		// Make a copy of the data so we can return the buffer to the pool
		result := make([]byte, length)
		copy(result, buf)
		mt.bufferPool.Put(buf)
		return result, nil
	}

	// For large messages, allocate a new buffer
	data := make([]byte, length)
	_, err := io.ReadFull(mt.reader, data)
	return data, err
}

// Close closes both the reader and writer.
func (mt *MsgpackTransport) Close() error {
	if err := mt.reader.Close(); err != nil {
		return err
	}
	return mt.writer.Close()
}

// Flush flushes the writer if it supports the Flush method.
func (mt *MsgpackTransport) Flush() error {
	if flusher, ok := mt.writer.(interface{ Flush() error }); ok {
		if err := flusher.Flush(); err != nil {
			return nil
		}
	}
	return nil
}
