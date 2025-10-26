package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type viewState int

const (
	SELECTING viewState = iota
	VIEWING
)

type endpoint struct {
	path   string
	method string
}

type model struct {
	choices           []endpoint
	search_endpoint   string
	selected_endpoint int // index in filtered list
	viewing_index     int // index in original choices slice
	state             viewState
}

// Init implements tea.Model.
func (m model) Init() tea.Cmd {
	return nil
}

func initialModel() model {
	return model{
		choices: []endpoint{
			{path: "/", method: "GET"},
			{path: "/", method: "POST"},
			{path: "/login", method: "POST"},
			{path: "/users", method: "GET"},
			{path: "/users", method: "POST"},
		},
		selected_endpoint: 0,
		search_endpoint:   "",
		state:             SELECTING,
	}
}

// filter endpoints based on search
func (m model) filteredChoices() []endpoint {
	if m.search_endpoint == "" {
		return m.choices
	}

	filtered := []endpoint{}
	for _, ep := range m.choices {
		if containsIgnoreCase(ep.path, m.search_endpoint) || containsIgnoreCase(ep.method, m.search_endpoint) {
			filtered = append(filtered, ep)
		}
	}
	return filtered
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			if m.state == VIEWING {
				m.state = SELECTING
			} else if m.state == SELECTING && len(m.search_endpoint) > 0 {
				m.search_endpoint = ""
			} else {
				return m, tea.Quit
			}

		case "up":
			if m.state == SELECTING && m.selected_endpoint > 0 {
				m.selected_endpoint--
			}

		case "down":
			if m.state == SELECTING && m.selected_endpoint < len(m.filteredChoices())-1 {
				m.selected_endpoint++
			}

		case "enter":
			if m.state == SELECTING {
				filtered := m.filteredChoices()
				if len(filtered) > 0 {
					selectedEp := filtered[m.selected_endpoint]
					// map to absolute index in original choices
					for i, ep := range m.choices {
						if ep == selectedEp {
							m.viewing_index = i
							break
						}
					}
				}
				m.state = VIEWING
			}

		case "backspace":
			if m.state == SELECTING && len(m.search_endpoint) > 0 {
				m.search_endpoint = m.search_endpoint[:len(m.search_endpoint)-1]
				m.selected_endpoint = 0
			}

		default:
			if m.state == SELECTING && len(msg.String()) == 1 {
				m.search_endpoint += msg.String()
				m.selected_endpoint = 0
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	s := "TAPI\n\n"

	switch m.state {
	case VIEWING:
		ep := m.choices[m.viewing_index]
		s += fmt.Sprintf("Method: %s\nPath: %s\n", ep.method, ep.path)
		s += "\nPress ESC to go back.\n"

	case SELECTING:
		s += fmt.Sprintf("Search: %s\n\n", m.search_endpoint)
		filtered := m.filteredChoices()
		if len(filtered) == 0 {
			s += "No matching endpoints.\n"
		} else {
			for i, ep := range filtered {
				cursor := " "
				if i == m.selected_endpoint {
					cursor = ">"
				}
				s += fmt.Sprintf("%s %s [%s]\n", cursor, ep.method, ep.path)
			}
		}
		s += "\nPress ESC to quit/clear search. Enter to view.\n"
	}

	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
