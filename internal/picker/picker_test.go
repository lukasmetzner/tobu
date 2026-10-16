package picker

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func testModel(t *testing.T, width, height int) *model {
	t.Helper()
	m := newModel([]Item{
		{Label: "/home/user/projects/tobu", Tag: "120.5"},
		{Label: "/home/user/projects/hcloud-go", Tag: "8.0"},
		{Label: "/home/user/.config/nvim", Tag: "0.2"},
		{Label: "scratch", Tag: "session", Accent: true},
	})
	m.Init()
	m.Update(tea.WindowSizeMsg{Width: width, Height: height})
	return m
}

func press(m *model, keys ...string) {
	for _, key := range keys {
		var msg tea.KeyPressMsg
		switch key {
		case "enter":
			msg = tea.KeyPressMsg{Code: tea.KeyEnter}
		case "esc":
			msg = tea.KeyPressMsg{Code: tea.KeyEscape}
		case "up":
			msg = tea.KeyPressMsg{Code: tea.KeyUp}
		case "down":
			msg = tea.KeyPressMsg{Code: tea.KeyDown}
		default:
			r := []rune(key)[0]
			msg = tea.KeyPressMsg{Code: r, Text: key}
		}
		m.Update(msg)
	}
}

func TestSelectAfterFiltering(t *testing.T) {
	m := testModel(t, 80, 24)
	press(m, "h", "c", "l", "enter")
	if !m.done || m.selected != 1 {
		t.Fatalf("done=%v selected=%d, want done and index 1", m.done, m.selected)
	}
}

func TestCursorCycles(t *testing.T) {
	m := testModel(t, 80, 24)
	press(m, "up", "enter")
	if m.selected != 3 {
		t.Fatalf("selected=%d, want 3", m.selected)
	}
}

func TestAbort(t *testing.T) {
	m := testModel(t, 80, 24)
	press(m, "down", "esc")
	if !m.done || m.selected != -1 {
		t.Fatalf("done=%v selected=%d, want done without selection", m.done, m.selected)
	}
}

func TestEnterWithoutMatchesIsIgnored(t *testing.T) {
	m := testModel(t, 80, 24)
	press(m, "x", "y", "z", "enter")
	if m.done {
		t.Fatal("expected picker to stay open without matches")
	}
}

func TestViewFitsTerminal(t *testing.T) {
	for _, width := range []int{12, 30, 80} {
		m := testModel(t, width, 20)
		lines := strings.Split(m.View().Content, "\n")
		if want := m.rows() + chromeHeight; len(lines) != want {
			t.Errorf("width %d: got %d lines, want %d", width, len(lines), want)
		}
		for _, line := range lines {
			if got := ansi.StringWidth(line); got != width {
				t.Errorf("width %d: line %q has width %d", width, ansi.Strip(line), got)
			}
		}
	}
}

func TestViewTruncatesPathsFromTheLeft(t *testing.T) {
	m := testModel(t, 30, 20)
	view := ansi.Strip(m.View().Content)
	if !strings.Contains(view, "…/projects/tobu") {
		t.Errorf("expected left-truncated path in view:\n%s", view)
	}
}

func TestViewIsClearedWhenDone(t *testing.T) {
	m := testModel(t, 80, 24)
	press(m, "enter")
	if content := m.View().Content; content != "" {
		t.Errorf("expected empty view after quitting, got %q", content)
	}
}
