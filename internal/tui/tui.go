package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

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
}

func (m modelUI) Init() tea.Cmd { return nil }

func (m modelUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
	b := strings.Builder{}
	b.WriteString("Grimmoir\n")
	b.WriteString("/ search  c toggle-stack  d delete  enter compose+copy  n new-from-clipboard  q quit\n\n")

	if m.searchMode {
		b.WriteString("Search: " + m.search + "\n\n")
	}
	if m.newMode {
		b.WriteString("New skill name: " + m.newName + "\n")
		b.WriteString("Press enter to save clipboard as new skill\n\n")
	}

	b.WriteString("Stack: [" + strings.Join(m.stack, ", ") + "]\n\n")

	b.WriteString("Skills\n")
	for i, s := range m.filtered {
		cursor := " "
		if i == m.selected {
			cursor = ">"
		}
		marker := " "
		if contains(m.stack, s.Name) {
			marker = "*"
		}
		b.WriteString(fmt.Sprintf("%s%s %s - %s\n", cursor, marker, s.Name, s.Description))
	}

	b.WriteString("\nPreview\n")
	if len(m.filtered) == 0 {
		b.WriteString("(no results)\n")
	} else {
		preview := strings.TrimSpace(m.filtered[m.selected].Body)
		if len(preview) > 600 {
			preview = preview[:600] + "..."
		}
		b.WriteString(preview + "\n")
	}

	if m.status != "" {
		b.WriteString("\nStatus: " + m.status + "\n")
	}

	return b.String()
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
