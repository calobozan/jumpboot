package jumpboot

import (
	"bufio"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// ExecOptions specifies a command to send to PythonExecProcess.
type ExecOptions struct {
	// ExecType is the command type: "exec" for code execution, "exit" to terminate.
	ExecType string `json:"type"`

	// Command is the Python code to execute (for "exec" type).
	Command string `json:"code"`
}

// ExecResult contains the response from PythonExecProcess.
type ExecResult struct {
	// ReturnType is "output" for success or "error" for exceptions.
	ReturnType string `json:"type"`

	// Output contains the result or error message.
	Output string `json:"output"`
}

//go:embed modules/pyprocexec/main.py
var pythonExecMain string

// PythonExecProcess provides simple command-based Python execution using JSON.
// Each Exec call sends code to Python and receives the result as JSON.
// This is simpler than QueueProcess but lacks bidirectional RPC capabilities.
type PythonExecProcess struct {
	*PythonProcess
}

// NewPythonExecProcess creates a Python process for simple command execution.
// Commands are sent as JSON and results are received as JSON responses.
func (env *Environment) NewPythonExecProcess(environment_vars map[string]string, extrafiles []*os.File) (*PythonExecProcess, error) {
	cwd, _ := os.Getwd()
	program := &PythonProgram{
		Name: "PythonExecProcess",
		Path: cwd,
		Program: Module{
			Name:   "__main__",
			Path:   filepath.Join(cwd, "modules", "main.py"),
			Source: base64.StdEncoding.EncodeToString([]byte(pythonExecMain)),
		},
		Modules:  []Module{},
		Packages: []Package{},
	}

	pyProcess, _, err := env.NewPythonProcessFromProgram(program, environment_vars, nil, false)
	if err != nil {
		return nil, err
	}

	return &PythonExecProcess{
		PythonProcess: pyProcess,
	}, nil
}

// Exec sends Python code for execution and returns the output.
// Returns an error if the code raised an exception or communication failed.
func (p *PythonExecProcess) Exec(code string) (string, error) {
	e := ExecOptions{
		ExecType: "exec",
		Command:  code,
	}

	// encode the command to JSON
	cmd_json, err := json.Marshal(e)
	if err != nil {
		return "", err
	}

	// send the command to the Python process
	_, err = p.PipeOut.Write([]byte(string(cmd_json) + "\n"))
	if err != nil {
		return "", err
	}

	// read the output from the Python process
	b, err := bufio.NewReader(p.PipeIn).ReadBytes('\n')
	if err != nil {
		return "", err
	}

	// decode the output from JSON
	var result ExecResult
	err = json.Unmarshal(b, &result)
	if err != nil {
		return "", err
	}

	if result.ReturnType == "error" {
		return "", errors.New(result.Output)
	} else {
		return result.Output, nil
	}
}

// Close sends an exit command to terminate the Python process.
func (p *PythonExecProcess) Close() {
	e := ExecOptions{
		ExecType: "exit",
		Command:  "",
	}

	// encode the command to JSON
	cmd_json, _ := json.Marshal(e)
	p.PipeOut.Write([]byte(string(cmd_json) + "\n"))
}
