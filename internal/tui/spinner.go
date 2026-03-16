package tui

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// spinnerFrames are the animation frames for the braille spinner.
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type spinnerMsg struct{}
type doneMsg struct{ result interface{} }
type errMsg struct{ err error }

// spinnerModel is the Bubbletea model for the spinner component.
type spinnerModel struct {
	label  string
	frame  int
	done   bool
	result interface{}
	err    error
}

func (m spinnerModel) Init() tea.Cmd {
	return tick()
}

func tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return spinnerMsg{}
	})
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinnerMsg:
		m.frame = (m.frame + 1) % len(spinnerFrames)
		return m, tick()
	case doneMsg:
		m.done = true
		m.result = msg.result
		return m, tea.Quit
	case errMsg:
		m.done = true
		m.err = msg.err
		return m, tea.Quit
	}
	return m, nil
}

func (m spinnerModel) View() string {
	if m.done {
		return ""
	}
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	return style.Render(spinnerFrames[m.frame]) + " " + m.label + "..."
}

// isStderrTTY returns true when stderr is an interactive terminal.
func isStderrTTY() bool {
	fi, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// RunWithSpinner runs fn in a goroutine and displays a spinner on stderr
// while it is running. The spinner is suppressed in non-TTY environments.
// Returns the value returned by fn.
func RunWithSpinner[T any](label string, fn func() (T, error)) (T, error) {
	if !isStderrTTY() {
		// No TTY — run directly without any UI.
		return fn()
	}

	type result struct {
		val T
		err error
	}
	ch := make(chan result, 1)

	model := spinnerModel{label: label}
	p := tea.NewProgram(model, tea.WithOutput(os.Stderr))

	go func() {
		// Wait 200ms before showing the spinner.
		time.Sleep(200 * time.Millisecond)
		v, err := fn()
		ch <- result{val: v, err: err}
		if err != nil {
			p.Send(errMsg{err: err})
		} else {
			p.Send(doneMsg{result: v})
		}
	}()

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "spinner error: %v\n", err)
	}

	res := <-ch
	return res.val, res.err
}
