// Package output provides structured response formatting with AI-native breadcrumbs.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// Mode controls output formatting.
type Mode int

const (
	ModeStyled Mode = iota // colored terminal (default for TTY)
	ModeJSON               // full JSON envelope
	ModeQuiet              // raw data only
	ModeAgent              // quiet + breadcrumbs
)

// Response is the standard output envelope (matches fizzy-cli pattern).
type Response struct {
	OK          bool         `json:"ok"`
	Data        any          `json:"data,omitempty"`
	Summary     string       `json:"summary,omitempty"`
	Breadcrumbs []Breadcrumb `json:"breadcrumbs,omitempty"`
	Error       string       `json:"error,omitempty"`
}

// Breadcrumb suggests a follow-up action to the agent or user.
type Breadcrumb struct {
	Action      string `json:"action"`
	Cmd         string `json:"cmd"`
	Description string `json:"description"`
}

// Printer handles formatted output.
type Printer struct {
	Out     io.Writer
	Err     io.Writer
	Mode    Mode
	Verbose bool
}

// DefaultPrinter returns a printer writing to stdout/stderr.
func DefaultPrinter() *Printer {
	return &Printer{
		Out:  os.Stdout,
		Err:  os.Stderr,
		Mode: ModeStyled,
	}
}

// Print outputs a successful response.
func (p *Printer) Print(resp Response) {
	resp.OK = true
	switch p.Mode {
	case ModeJSON, ModeAgent:
		p.printJSON(resp)
	case ModeQuiet:
		if resp.Summary != "" {
			fmt.Fprintln(p.Out, resp.Summary)
		}
	default:
		p.printStyled(resp)
	}
}

// PrintError outputs an error response.
func (p *Printer) PrintError(msg string) {
	resp := Response{OK: false, Error: msg}
	switch p.Mode {
	case ModeJSON, ModeAgent:
		p.printJSON(resp)
	default:
		fmt.Fprintf(p.Err, "\033[31merror:\033[0m %s\n", msg)
	}
}

// Info prints an informational message (styled mode only).
func (p *Printer) Info(format string, args ...any) {
	if p.Mode == ModeStyled {
		fmt.Fprintf(p.Out, "\033[32m✓\033[0m %s\n", fmt.Sprintf(format, args...))
	}
}

// Warn prints a warning message.
func (p *Printer) Warn(format string, args ...any) {
	if p.Mode == ModeStyled {
		fmt.Fprintf(p.Err, "\033[33m⚠\033[0m %s\n", fmt.Sprintf(format, args...))
	}
}

// Debug prints a verbose message (only when verbose is enabled).
func (p *Printer) Debug(format string, args ...any) {
	if p.Verbose {
		fmt.Fprintf(p.Err, "\033[90m  %s\033[0m\n", fmt.Sprintf(format, args...))
	}
}

func (p *Printer) printJSON(resp Response) {
	enc := json.NewEncoder(p.Out)
	enc.SetIndent("", "  ")
	_ = enc.Encode(resp)
}

func (p *Printer) printStyled(resp Response) {
	if resp.Summary != "" {
		fmt.Fprintln(p.Out, resp.Summary)
	}
	if len(resp.Breadcrumbs) > 0 {
		fmt.Fprintln(p.Out)
		fmt.Fprintln(p.Out, "\033[90mNext:\033[0m")
		for _, b := range resp.Breadcrumbs {
			fmt.Fprintf(p.Out, "  \033[36m%s\033[0m  %s\n", b.Cmd, b.Description)
		}
	}
}

// ParseMode converts flag values to a Mode.
func ParseMode(jsonFlag, quietFlag, agentFlag bool) Mode {
	switch {
	case jsonFlag:
		return ModeJSON
	case agentFlag:
		return ModeAgent
	case quietFlag:
		return ModeQuiet
	default:
		return ModeStyled
	}
}

// Table prints a simple aligned table.
func (p *Printer) Table(rows [][]string) {
	if len(rows) == 0 {
		return
	}
	// Calculate column widths.
	widths := make([]int, len(rows[0]))
	for _, row := range rows {
		for i, col := range row {
			if i < len(widths) && len(col) > widths[i] {
				widths[i] = len(col)
			}
		}
	}
	for _, row := range rows {
		parts := make([]string, len(row))
		for i, col := range row {
			if i < len(widths) {
				parts[i] = fmt.Sprintf("%-*s", widths[i], col)
			} else {
				parts[i] = col
			}
		}
		fmt.Fprintln(p.Out, strings.Join(parts, "  "))
	}
}
