package jumpboot

import (
	"encoding/binary"
	"io"

	"github.com/vmihailenco/msgpack/v5"
)

type MsgpackSerializer struct{}

func (ms MsgpackSerializer) Marshal(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

func (ms MsgpackSerializer) Unmarshal(data []byte, v interface{}) error {
	return msgpack.Unmarshal(data, v)
}

type MsgpackTransport struct {
	reader     io.ReadCloser
	writer     io.WriteCloser
	bufferPool *BufferPool
}

func NewMsgpackTransport(reader io.ReadCloser, writer io.WriteCloser) *MsgpackTransport {
	return &MsgpackTransport{reader: reader,
		writer:     writer,
		bufferPool: NewBufferPool(8192, 10), // Same size as Python side
	}
}

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

func (mt *MsgpackTransport) Close() error {
	if err := mt.reader.Close(); err != nil {
		return err
	}
	return mt.writer.Close()
}

func (mt *MsgpackTransport) Flush() error {
	if flusher, ok := mt.writer.(interface{ Flush() error }); ok {
		if err := flusher.Flush(); err != nil {
			return nil
		}
	}
	return nil
}
