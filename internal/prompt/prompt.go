package prompt

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var builtinHelp = `
FuseMomo Interactive Prompt
════════════════════════════════════════
Type a subcommand (with or without 'fusemomo' prefix) and press Enter.
Commands are executed directly and output is shown inline.

  entity resolve --id <key>=<value>   Resolve identifiers
  entity get <uuid>                   Get entity profile
  entity list                         List entities
  entity delete <uuid> --confirm      GDPR anonymize
  entity link --entity-id <uuid> ...  Link identifiers
  interaction log --entity-id ...     Log interaction
  interaction batch --file <path>     Batch log
  recommend get --entity-id ...       Get recommendation
  recommend outcome <rec-id> ...      Record outcome
  version                             Print build info
  help                                Show this help
  exit / quit / Ctrl+D                Leave the REPL
════════════════════════════════════════
`

// Run starts the interactive REPL loop.
// Terminal state is never modified — bufio reads in cooked mode.
func Run() {
	binary := os.Args[0] // path to the fusemomo binary itself

	fmt.Print(builtinHelp)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\033[36mfusemomo\033[0m > ")

		if !scanner.Scan() {
			// EOF (Ctrl+D) or error — exit cleanly.
			fmt.Println()
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Strip optional leading "fusemomo " prefix.
		line = strings.TrimPrefix(line, "fusemomo ")
		line = strings.TrimSpace(line)

		switch line {
		case "exit", "quit":
			fmt.Println("Goodbye, See You Again!")
			return
		case "help":
			fmt.Print(builtinHelp)
			continue
		case "prompt":
			fmt.Fprintln(os.Stderr, "You are already in the interactive prompt.")
			continue
		}

		// Parse the line into args and run the binary with them.
		args, err := splitArgs(line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  error parsing command: %v\n", err)
			continue
		}

		cmd := exec.Command(binary, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			// Non-zero exit from a subcommand is expected (e.g. 404).
			// The subcommand already printed its own error — don't double-print.
			if exitErr, ok := err.(*exec.ExitError); ok {
				_ = exitErr // exit code already handled by the subcommand
			}
		}
	}
}

// splitArgs splits a command line into tokens, respecting quoted strings.
func splitArgs(line string) ([]string, error) {
	var args []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for _, r := range line {
		switch {
		case inQuote && r == quoteChar:
			inQuote = false
		case !inQuote && (r == '"' || r == '\''):
			inQuote = true
			quoteChar = r
		case !inQuote && r == ' ':
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if inQuote {
		return nil, fmt.Errorf("unclosed quote in: %s", line)
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args, nil
}
