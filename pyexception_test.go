package jumpboot

import (
	"strings"
	"testing"
)

func TestPythonExceptionFromJSON(t *testing.T) {
	jsonData := []byte(`{
		"exception": "ValueError",
		"message": "invalid value",
		"traceback": "Traceback (most recent call last):\n  File \"test.py\", line 1\nValueError: invalid value"
	}`)

	ex, err := NewPythonExceptionFromJSON(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse exception: %v", err)
	}

	if ex.Exception != "ValueError" {
		t.Errorf("Expected exception type 'ValueError', got '%s'", ex.Exception)
	}
	if ex.Message != "invalid value" {
		t.Errorf("Expected message 'invalid value', got '%s'", ex.Message)
	}
	if ex.Cause != nil {
		t.Error("Expected Cause to be nil for simple exception")
	}
}

func TestPythonExceptionWithCause(t *testing.T) {
	jsonData := []byte(`{
		"exception": "RuntimeError",
		"message": "operation failed",
		"traceback": "Traceback (most recent call last):\nRuntimeError: operation failed",
		"cause": {
			"exception": "IOError",
			"message": "file not found",
			"traceback": "Traceback (most recent call last):\nIOError: file not found"
		}
	}`)

	ex, err := NewPythonExceptionFromJSON(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse exception with cause: %v", err)
	}

	if ex.Exception != "RuntimeError" {
		t.Errorf("Expected exception type 'RuntimeError', got '%s'", ex.Exception)
	}
	if ex.Cause == nil {
		t.Fatal("Expected Cause to be non-nil for chained exception")
	}
	if ex.Cause.Exception != "IOError" {
		t.Errorf("Expected cause exception type 'IOError', got '%s'", ex.Cause.Exception)
	}
	if ex.Cause.Message != "file not found" {
		t.Errorf("Expected cause message 'file not found', got '%s'", ex.Cause.Message)
	}
}

func TestPythonExceptionWithArgs(t *testing.T) {
	jsonData := []byte(`{
		"exception": "OSError",
		"message": "[Errno 2] No such file or directory: 'test.txt'",
		"traceback": "Traceback...",
		"args": [2, "No such file or directory", "test.txt"]
	}`)

	ex, err := NewPythonExceptionFromJSON(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse exception with args: %v", err)
	}

	if len(ex.ExceptionArgs) != 3 {
		t.Fatalf("Expected 3 args, got %d", len(ex.ExceptionArgs))
	}

	// First arg should be errno (float64 from JSON)
	if errno, ok := ex.ExceptionArgs[0].(float64); !ok || errno != 2 {
		t.Errorf("Expected errno 2, got %v", ex.ExceptionArgs[0])
	}
}

func TestPythonExceptionToStringWithCause(t *testing.T) {
	ex := &PythonException{
		Exception: "RuntimeError",
		Message:   "top level error",
		Traceback: "Traceback...",
		Cause: &PythonException{
			Exception: "ValueError",
			Message:   "underlying error",
			Traceback: "Inner traceback...",
		},
	}

	str := ex.ToString()
	if !strings.Contains(str, "RuntimeError") {
		t.Error("ToString should contain RuntimeError")
	}
	if !strings.Contains(str, "Caused by:") {
		t.Error("ToString should contain 'Caused by:' for chained exceptions")
	}
	if !strings.Contains(str, "ValueError") {
		t.Error("ToString should contain the cause exception type")
	}
}

func TestPythonExceptionError(t *testing.T) {
	ex := &PythonException{
		Exception: "KeyError",
		Message:   "'missing_key'",
		Traceback: "Traceback...",
	}

	err := ex.Error()
	if err == nil {
		t.Fatal("Error() should return non-nil error")
	}
	if !strings.Contains(err.Error(), "KeyError") {
		t.Error("Error string should contain exception type")
	}
}

func TestPythonExceptionNestedCause(t *testing.T) {
	// Test deeply nested exception chain
	jsonData := []byte(`{
		"exception": "ApplicationError",
		"message": "application failed",
		"traceback": "TB1",
		"cause": {
			"exception": "DatabaseError",
			"message": "database failed",
			"traceback": "TB2",
			"cause": {
				"exception": "ConnectionError",
				"message": "connection refused",
				"traceback": "TB3"
			}
		}
	}`)

	ex, err := NewPythonExceptionFromJSON(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse nested exception: %v", err)
	}

	if ex.Exception != "ApplicationError" {
		t.Errorf("Expected top-level exception 'ApplicationError', got '%s'", ex.Exception)
	}
	if ex.Cause == nil || ex.Cause.Exception != "DatabaseError" {
		t.Error("Expected first-level cause to be DatabaseError")
	}
	if ex.Cause.Cause == nil || ex.Cause.Cause.Exception != "ConnectionError" {
		t.Error("Expected second-level cause to be ConnectionError")
	}

	str := ex.ToString()
	if !strings.Contains(str, "ApplicationError") ||
		!strings.Contains(str, "DatabaseError") ||
		!strings.Contains(str, "ConnectionError") {
		t.Error("ToString should include all chained exceptions")
	}
}
