package picker

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

var ErrAborted = errors.New("selection aborted")

type Item struct {
	Label  string
	Tag    string
	Accent bool
}

const (
	heightRatio  = 0.45
	minRows      = 3
	chromeHeight = 3
	chromeWidth  = 4
	pointer      = "▌ "
	noPointer    = "  "
	ellipsis     = "…"
)

var (
	borderStyle   = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("8")).Padding(0, 1)
	promptStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Bold(true)
	counterStyle  = lipgloss.NewStyle().Faint(true)
	pointerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	tagStyle      = lipgloss.NewStyle().Faint(true)
	accentStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	labelStyle    = lipgloss.NewStyle()
	selectedStyle = lipgloss.NewStyle().Bold(true)
	matchColor    = lipgloss.Color("6")
)

func Run(ctx context.Context, items []Item) (int, error) {
	m := newModel(items)
	if _, err := tea.NewProgram(m, tea.WithContext(ctx)).Run(); err != nil {
		return -1, err
	}
	if m.selected < 0 {
		return -1, ErrAborted
	}
	return m.selected, nil
}

type model struct {
	items    []Item
	tags     []string
	input    textinput.Model
	matches  []match
	cursor   int
	offset   int
	width    int
	height   int
	selected int
	done     bool
}

func newModel(items []Item) *model {
	input := textinput.New()
	input.Prompt = "> "
	styles := input.Styles()
	styles.Focused.Prompt = promptStyle
	input.SetStyles(styles)

	m := &model{
		items:    items,
		tags:     alignTags(items),
		input:    input,
		selected: -1,
	}
	m.filter()
	return m
}

func (m *model) Init() tea.Cmd {
	return m.input.Focus()
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.input.SetWidth(m.contentWidth() - ansi.StringWidth(m.input.Prompt) - m.counterWidth() - 2)
		m.scroll()
		return m, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+g", "ctrl+q", "esc":
			return m.quit()
		case "enter":
			if len(m.matches) == 0 {
				return m, nil
			}
			m.selected = m.matches[m.cursor].index
			return m.quit()
		case "up", "ctrl+p", "ctrl+k", "shift+tab":
			m.cycle(-1)
			return m, nil
		case "down", "ctrl+n", "ctrl+j", "tab":
			m.cycle(1)
			return m, nil
		case "pgup":
			m.jump(-m.rows())
			return m, nil
		case "pgdown":
			m.jump(m.rows())
			return m, nil
		}
	}

	query := m.input.Value()
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	if m.input.Value() != query {
		m.filter()
	}
	return m, cmd
}

func (m *model) View() tea.View {
	if m.done || m.width == 0 {
		return tea.NewView("")
	}

	width := m.contentWidth()
	rows := m.rows()
	lines := make([]string, 0, rows+1)
	lines = append(lines, m.promptLine(width))
	for i := range rows {
		lines = append(lines, m.matchLine(m.offset+i, width))
	}
	return tea.NewView(borderStyle.Render(strings.Join(lines, "\n")))
}

func (m *model) quit() (tea.Model, tea.Cmd) {
	m.done = true
	return m, tea.Quit
}

func (m *model) filter() {
	m.matches = filter(m.items, m.input.Value())
	m.cursor = 0
	m.offset = 0
}

func (m *model) cycle(delta int) {
	if len(m.matches) == 0 {
		return
	}
	m.cursor = (m.cursor + delta + len(m.matches)) % len(m.matches)
	m.scroll()
}

func (m *model) jump(delta int) {
	if len(m.matches) == 0 {
		return
	}
	m.cursor = min(max(m.cursor+delta, 0), len(m.matches)-1)
	m.scroll()
}

func (m *model) scroll() {
	rows := m.rows()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+rows {
		m.offset = m.cursor - rows + 1
	}
}

func (m *model) rows() int {
	rows := max(int(float64(m.height)*heightRatio)-chromeHeight, minRows)
	return max(min(rows, m.height-chromeHeight), 1)
}

func (m *model) contentWidth() int {
	return max(m.width-chromeWidth, 1)
}

func (m *model) counterWidth() int {
	return 2*len(strconv.Itoa(len(m.items))) + 1
}

func (m *model) promptLine(width int) string {
	input := m.input.View()
	counter := counterStyle.Render(fmt.Sprintf("%d/%d", len(m.matches), len(m.items)))
	gap := max(width-ansi.StringWidth(input)-ansi.StringWidth(counter), 1)
	return fit(input+strings.Repeat(" ", gap)+counter, width)
}

func (m *model) matchLine(i, width int) string {
	if i >= len(m.matches) {
		return fit("", width)
	}

	match := m.matches[i]
	item := m.items[match.index]
	style, prefix := labelStyle, noPointer
	if i == m.cursor {
		style, prefix = selectedStyle, pointerStyle.Render(pointer)
	}
	tag := tagStyle
	if item.Accent {
		tag = accentStyle
	}
	prefix += tag.Render(m.tags[match.index]) + "  "

	label := highlight(item.Label, match.positions, style)
	available := max(width-ansi.StringWidth(prefix), 1)
	if excess := ansi.StringWidth(label) - available; excess > 0 {
		label = ansi.TruncateLeft(label, excess+ansi.StringWidth(ellipsis), ellipsis)
	}
	return fit(prefix+label, width)
}

func highlight(text string, positions []int, style lipgloss.Style) string {
	runes := []rune(text)
	marked := make([]bool, len(runes))
	for _, position := range positions {
		marked[position] = true
	}

	matchStyle := style.Foreground(matchColor)
	var b strings.Builder
	for start := 0; start < len(runes); {
		end := start + 1
		for end < len(runes) && marked[end] == marked[start] {
			end++
		}
		segmentStyle := style
		if marked[start] {
			segmentStyle = matchStyle
		}
		b.WriteString(segmentStyle.Render(string(runes[start:end])))
		start = end
	}
	return b.String()
}

func fit(line string, width int) string {
	line = ansi.Truncate(line, width, "")
	return line + strings.Repeat(" ", width-ansi.StringWidth(line))
}

func alignTags(items []Item) []string {
	width := 0
	for _, item := range items {
		width = max(width, ansi.StringWidth(item.Tag))
	}
	tags := make([]string, len(items))
	for i, item := range items {
		tags[i] = strings.Repeat(" ", width-ansi.StringWidth(item.Tag)) + item.Tag
	}
	return tags
}
