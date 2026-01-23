package jumpboot

import (
	"encoding/json"
	"fmt"
)

// PythonException represents an exception raised in a Python process.
// It captures the exception type, message, and full traceback for debugging.
type PythonException struct {
	// Exception is the exception class name (e.g., "ValueError", "KeyError").
	Exception string `json:"exception"`

	// Message is the exception message/description.
	Message string `json:"message"`

	// Traceback is the full Python traceback string.
	Traceback string `json:"traceback"`
}

// ToString formats the exception as a readable string with type, message, and traceback.
func (e *PythonException) ToString() string {
	return fmt.Sprintf("%s: %s\n%s", e.Exception, e.Message, e.Traceback)
}

// Error returns the exception as a Go error.
func (e *PythonException) Error() error {
	return fmt.Errorf("%s", e.ToString())
}

// NewPythonExceptionFromJSON parses a PythonException from JSON bytes.
// This is used to deserialize exceptions sent from Python via the status pipe.
func NewPythonExceptionFromJSON(data []byte) (*PythonException, error) {
	var pyException PythonException
	err := json.Unmarshal(data, &pyException)
	if err != nil {
		return nil, err
	}
	return &pyException, nil
}
