package jumpboot

type Serializer interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

type Transport interface {
	Send(data []byte) error
	Receive() ([]byte, error)
	Close() error
	Flush() error
}
