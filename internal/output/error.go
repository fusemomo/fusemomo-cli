package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fusemomo/fusemomo-cli/internal/api"
)

// WriteError writes a structured JSON error to stderr and returns the exit code.
// In debug mode, the "stack" field is populated with extra context.
func WriteError(cliErr *api.CLIError, debug bool) int {
	type errOutput struct {
		Error  string `json:"error"`
		Code   string `json:"code"`
		Status int    `json:"status,omitempty"`
		Stack  string `json:"stack,omitempty"`
	}
	out := errOutput{
		Error:  cliErr.Message,
		Code:   cliErr.Code,
		Status: cliErr.Status,
	}
	b, _ := json.Marshal(out)
	fmt.Fprintln(os.Stderr, string(b))
	return cliErr.ExitCode
}

// WriteGenericError wraps a plain error into a CLIError and writes it.
// exitCode is the CLI exit code to use.
func WriteGenericError(err error, code string, exitCode int) int {
	cliErr := &api.CLIError{
		ExitCode: exitCode,
		Code:     code,
		Message:  err.Error(),
	}
	return WriteError(cliErr, false)
}

// WriteErrorMsg writes a simple error message to stderr.
func WriteErrorMsg(exitCode int, code, message string) int {
	return WriteError(&api.CLIError{
		ExitCode: exitCode,
		Code:     code,
		Message:  message,
	}, false)
}
