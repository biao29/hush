package output

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestPrintJSON(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Out: &buf, Err: &buf, Mode: ModeJSON}

	p.Print(Response{
		Data:    map[string]string{"project": "test"},
		Summary: "1 project synced",
		Breadcrumbs: []Breadcrumb{
			{Action: "edit", Cmd: "hush edit test", Description: "Edit rules"},
		},
	})

	var resp Response
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if !resp.OK {
		t.Error("expected ok=true")
	}
	if resp.Summary != "1 project synced" {
		t.Errorf("summary = %q, want %q", resp.Summary, "1 project synced")
	}
	if len(resp.Breadcrumbs) != 1 {
		t.Errorf("breadcrumbs count = %d, want 1", len(resp.Breadcrumbs))
	}
}

func TestPrintErrorJSON(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Out: &buf, Err: &buf, Mode: ModeJSON}

	p.PrintError("something failed")

	var resp Response
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp.OK {
		t.Error("expected ok=false")
	}
	if resp.Error != "something failed" {
		t.Errorf("error = %q, want %q", resp.Error, "something failed")
	}
}

func TestPrintStyled(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Out: &buf, Err: &buf, Mode: ModeStyled}

	p.Print(Response{Summary: "all good"})

	if !bytes.Contains(buf.Bytes(), []byte("all good")) {
		t.Error("styled output should contain summary")
	}
}

func TestPrintQuiet(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Out: &buf, Err: &buf, Mode: ModeQuiet}

	p.Print(Response{
		Summary: "done",
		Breadcrumbs: []Breadcrumb{
			{Action: "next", Cmd: "hush sync", Description: "Sync"},
		},
	})

	out := buf.String()
	if out != "done\n" {
		t.Errorf("quiet output = %q, want %q", out, "done\n")
	}
}

func TestParseMode(t *testing.T) {
	tests := []struct {
		json, quiet, agent bool
		want               Mode
	}{
		{false, false, false, ModeStyled},
		{true, false, false, ModeJSON},
		{false, true, false, ModeQuiet},
		{false, false, true, ModeAgent},
		{true, true, false, ModeJSON}, // json takes precedence
	}
	for _, tt := range tests {
		got := ParseMode(tt.json, tt.quiet, tt.agent)
		if got != tt.want {
			t.Errorf("ParseMode(%v,%v,%v) = %d, want %d", tt.json, tt.quiet, tt.agent, got, tt.want)
		}
	}
}

func TestTable(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Out: &buf, Err: &buf, Mode: ModeStyled}

	p.Table([][]string{
		{"name", "status"},
		{"project-a", "synced"},
		{"project-bbb", "stale"},
	})

	out := buf.String()
	if !bytes.Contains([]byte(out), []byte("project-a")) {
		t.Error("table should contain project-a")
	}
	if !bytes.Contains([]byte(out), []byte("project-bbb")) {
		t.Error("table should contain project-bbb")
	}
}
