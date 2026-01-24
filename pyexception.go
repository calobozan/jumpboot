package jumpboot

import (
	"encoding/json"
	"fmt"
)

// PythonException represents an exception raised in a Python process.
// It captures the exception type, message, full traceback, and optional
// chained exception for debugging.
type PythonException struct {
	// Exception is the exception class name (e.g., "ValueError", "KeyError").
	Exception string `json:"exception"`

	// Message is the exception message/description.
	Message string `json:"message"`

	// Traceback is the full Python traceback string.
	Traceback string `json:"traceback"`

	// Cause is the chained exception from "raise X from Y" syntax.
	// This field is nil if there is no chained exception.
	Cause *PythonException `json:"cause,omitempty"`

	// ExceptionArgs contains the structured arguments passed to the exception
	// constructor, if available. For example, OSError may include errno and filename.
	ExceptionArgs []interface{} `json:"args,omitempty"`
}

// ToString formats the exception as a readable string with type, message, and traceback.
// If a chained exception (Cause) exists, it is included in the output.
func (e *PythonException) ToString() string {
	result := fmt.Sprintf("%s: %s\n%s", e.Exception, e.Message, e.Traceback)
	if e.Cause != nil {
		result += fmt.Sprintf("\n\nCaused by:\n%s", e.Cause.ToString())
	}
	return result
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
