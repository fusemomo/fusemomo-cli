package output

import (
	"encoding/json"
	"fmt"
	"os"
)

// IsTTY returns true if the given file is an interactive terminal.
func IsTTY(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// JSON writes v to stdout.
// If stdout is a TTY: pretty-printed with 2-space indent.
// If stdout is piped: compact single-line JSON (machine consumption).
func JSON(v interface{}) error {
	var b []byte
	var err error
	if IsTTY(os.Stdout) {
		b, err = json.MarshalIndent(v, "", "  ")
	} else {
		b, err = json.Marshal(v)
	}
	if err != nil {
		return fmt.Errorf("failed to marshal output: %w", err)
	}
	fmt.Println(string(b))
	return nil
}
