package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"grimmoir/internal/model"
	"grimmoir/internal/service"
)

type App struct {
	svc *service.Service
}

func New(svc *service.Service) *App {
	return &App{svc: svc}
}

func (a *App) Run() error {
	skills, err := a.svc.List()
	if err != nil {
		return err
	}
	m := modelUI{
		svc:      a.svc,
		all:      skills,
		filtered: skills,
		selected: 0,
		stack:    make([]string, 0),
	}
	_, err = tea.NewProgram(m).Run()
	return err
}

type modelUI struct {
	svc *service.Service

	all      []model.Skill
	filtered []model.Skill
	selected int
	stack    []string

	searchMode bool
	search     string

	newMode bool
	newName string

	status string

	width  int
	height int
}

func (m modelUI) Init() tea.Cmd { return nil }

func (m modelUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		key := msg.String()

		if m.searchMode {
			switch key {
			case "esc", "enter":
				m.searchMode = false
				return m, nil
			case "backspace":
				if len(m.search) > 0 {
					m.search = m.search[:len(m.search)-1]
					m.applySearch()
				}
				return m, nil
			default:
				if len(key) == 1 {
					m.search += key
					m.applySearch()
				}
				return m, nil
			}
		}

		if m.newMode {
			switch key {
			case "esc":
				m.newMode = false
				m.newName = ""
				m.status = ""
				return m, nil
			case "enter":
				path, err := m.svc.AddFromClipboard(model.Skill{Name: strings.TrimSpace(m.newName)})
				if err != nil {
					m.status = "Add failed: " + err.Error()
					return m, nil
				}
				skills, err := m.svc.List()
				if err == nil {
					m.all = skills
					m.applySearch()
				}
				m.newMode = false
				m.newName = ""
				m.status = "Saved " + path
				return m, nil
			case "backspace":
				if len(m.newName) > 0 {
					m.newName = m.newName[:len(m.newName)-1]
				}
				return m, nil
			default:
				if len(key) == 1 {
					m.newName += key
				}
				return m, nil
			}
		}

		switch key {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "/":
			m.searchMode = true
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.filtered)-1 {
				m.selected++
			}
		case "c":
			if len(m.filtered) > 0 {
				name := m.filtered[m.selected].Name
				if contains(m.stack, name) {
					m.stack = remove(m.stack, name)
				} else {
					m.stack = append(m.stack, name)
				}
			}
		case "enter":
			if len(m.stack) == 0 {
				break
			}
			text, err := m.svc.Compose(m.stack)
			if err != nil {
				m.status = "Compose failed: " + err.Error()
				break
			}
			if err := m.svc.CopyToClipboard(text); err != nil {
				m.status = "Copy failed: " + err.Error()
				break
			}
			m.status = fmt.Sprintf("Copied %d skills to clipboard", len(m.stack))
		case "n":
			m.newMode = true
			m.newName = ""
			m.status = ""
		case "d":
			if len(m.filtered) == 0 {
				break
			}
			name := m.filtered[m.selected].Name
			path, err := m.svc.DeleteByName(name)
			if err != nil {
				m.status = "Delete failed: " + err.Error()
				break
			}
			m.stack = remove(m.stack, name)
			skills, err := m.svc.List()
			if err != nil {
				m.status = "Refresh failed: " + err.Error()
				break
			}
			m.all = skills
			m.applySearch()
			m.status = "Deleted " + path
		}
	}

	return m, nil
}

func (m modelUI) View() string {
	styles := newStyles()

	header := styles.header.Render("Grimmoir Prompt Studio")
	contentWidth := 104
	if m.width > 0 {
		contentWidth = max(28, m.width-6)
	}
	commands := styles.panel.Width(contentWidth).Render(m.renderCommandGuide(styles))

	stateRows := make([]string, 0, 2)
	if m.searchMode {
		stateRows = append(stateRows, styles.badge.Render("Search: "+m.search))
	}
	if m.newMode {
		stateRows = append(stateRows, styles.badge.Render("New skill name: "+m.newName+" (enter to save clipboard)"))
	}
	state := ""
	if len(stateRows) > 0 {
		state = strings.Join(stateRows, "\n") + "\n"
	}

	stack := m.renderStack(styles, contentWidth)

	maxListRows := 10
	maxPreviewLines := 28
	if m.height > 0 {
		maxListRows = max(4, (m.height-18)/2)
		maxPreviewLines = max(6, m.height-16)
	}

	browser := m.renderList(styles, contentWidth, maxListRows)
	preview := m.renderPreview(styles, contentWidth, maxPreviewLines)

	body := ""
	if contentWidth < 92 {
		body = lipgloss.JoinVertical(
			lipgloss.Left,
			styles.panel.Width(contentWidth).Render(stack),
			styles.panel.Width(contentWidth).Render(browser),
			styles.panel.Width(contentWidth).Render(preview),
		)
	} else {
		leftWidth := max(34, contentWidth/2-1)
		rightWidth := max(34, contentWidth-leftWidth-1)
		left := lipgloss.JoinVertical(lipgloss.Left,
			styles.panel.Width(leftWidth).Render(stack),
			styles.panel.Width(leftWidth).Render(browser),
		)
		right := styles.panel.Width(rightWidth).Render(preview)
		body = lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	}

	footer := ""
	if m.status != "" {
		footer = "\n" + styles.status.Width(contentWidth).Render(truncate(m.status, contentWidth-2))
	}

	return styles.base.Render(header + "\n\n" + commands + "\n\n" + state + body + footer)
}

func (m modelUI) renderCommandGuide(styles uiStyles) string {
	searchState := "OFF"
	if m.searchMode {
		searchState = "ON"
	}
	newState := "OFF"
	if m.newMode {
		newState = "ON"
	}
	composeState := "READY"
	if len(m.stack) == 0 {
		composeState = "NEED STACK"
	}
	deleteState := "READY"
	if len(m.filtered) == 0 {
		deleteState = "NO SELECTION"
	}

	b := strings.Builder{}
	b.WriteString(styles.sectionTitle.Render("Command Guide"))
	b.WriteString("\n")
	b.WriteString(styles.hint.Render("search [/]: " + searchState + "   new [n]: " + newState))
	b.WriteString("\n")
	b.WriteString(styles.hint.Render("move [j/k or arrows]   stack [c]   delete [d]: " + deleteState + "   enter compose+copy: " + composeState + "   quit [q]"))
	return b.String()
}

func (m modelUI) renderStack(styles uiStyles, panelWidth int) string {
	if len(m.stack) == 0 {
		return styles.sectionTitle.Render("Prompt Stack") + "\n" + styles.muted.Render("No prompts selected")
	}
	tokens := make([]string, 0, len(m.stack))
	for _, name := range m.stack {
		tokens = append(tokens, styles.pill.Render(truncate(name, max(8, panelWidth/3))))
	}
	return styles.sectionTitle.Render("Prompt Stack") + "\n" + strings.Join(tokens, " ")
}

func (m modelUI) renderList(styles uiStyles, panelWidth int, maxRows int) string {
	b := strings.Builder{}
	b.WriteString(styles.sectionTitle.Render("Skill Browser"))
	b.WriteString("\n")
	if len(m.filtered) == 0 {
		b.WriteString(styles.muted.Render("No results"))
		return b.String()
	}
	end := len(m.filtered)
	if end > maxRows {
		end = maxRows
	}
	for i := 0; i < end; i++ {
		s := m.filtered[i]
		row := fmt.Sprintf("%s  %s", s.Name, s.Description)
		row = truncate(row, max(8, panelWidth-8))
		if contains(m.stack, s.Name) {
			row = "* " + row
		} else {
			row = "  " + row
		}
		if i == m.selected {
			b.WriteString(styles.selected.Render("▸ " + row))
		} else {
			b.WriteString(styles.row.Render("  " + row))
		}
		if i < end-1 {
			b.WriteString("\n")
		}
	}
	if len(m.filtered) > end {
		b.WriteString("\n" + styles.muted.Render(fmt.Sprintf("... %d more", len(m.filtered)-end)))
	}
	return b.String()
}

func (m modelUI) renderPreview(styles uiStyles, panelWidth int, maxLines int) string {
	b := strings.Builder{}
	b.WriteString(styles.sectionTitle.Render("Preview Pane"))
	b.WriteString("\n")
	if len(m.filtered) == 0 {
		b.WriteString(styles.muted.Render("(no results)"))
		return b.String()
	}
	preview := strings.TrimSpace(m.filtered[m.selected].Body)
	lines := strings.Split(preview, "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines = append(lines, "...")
	}
	for i := range lines {
		lines[i] = truncate(lines[i], max(10, panelWidth-6))
	}
	b.WriteString(styles.preview.Render(strings.Join(lines, "\n")))
	return b.String()
}

type uiStyles struct {
	base         lipgloss.Style
	header       lipgloss.Style
	hint         lipgloss.Style
	badge        lipgloss.Style
	panel        lipgloss.Style
	sectionTitle lipgloss.Style
	row          lipgloss.Style
	selected     lipgloss.Style
	pill         lipgloss.Style
	preview      lipgloss.Style
	status       lipgloss.Style
	muted        lipgloss.Style
}

func newStyles() uiStyles {
	return uiStyles{
		base:         lipgloss.NewStyle().Padding(1, 2),
		header:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("24")).Padding(0, 1),
		hint:         lipgloss.NewStyle().Foreground(lipgloss.Color("109")),
		badge:        lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("60")).Padding(0, 1),
		panel:        lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62")).Padding(1, 1).MarginRight(1),
		sectionTitle: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("221")),
		row:          lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		selected:     lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("31")).Bold(true),
		pill:         lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("28")).Padding(0, 1),
		preview:      lipgloss.NewStyle().Foreground(lipgloss.Color("250")),
		status:       lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("22")).Padding(0, 1),
		muted:        lipgloss.NewStyle().Foreground(lipgloss.Color("244")),
	}
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func truncate(s string, width int) string {
	if width <= 0 || len(s) <= width {
		return s
	}
	if width <= 3 {
		return s[:width]
	}
	return s[:width-3] + "..."
}

func (m *modelUI) applySearch() {
	m.filtered = service.Search(m.all, m.search)
	if len(m.filtered) == 0 {
		m.selected = 0
		return
	}
	if m.selected >= len(m.filtered) {
		m.selected = len(m.filtered) - 1
	}
}

func contains(values []string, target string) bool {
	for _, v := range values {
		if strings.EqualFold(v, target) {
			return true
		}
	}
	return false
}

func remove(values []string, target string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		if strings.EqualFold(v, target) {
			continue
		}
		out = append(out, v)
	}
	return out
}
