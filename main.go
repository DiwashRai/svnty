package main

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
)

// Not to be confused with a bubbletea 'model'
type SvnModel struct {
	workingPath string
	remoteUrl   string
	revision    uint32
}

type AppModel struct {
	svnModel SvnModel
}

func New() AppModel {
	return AppModel{}
}

func (m AppModel) Init() tea.Cmd {
	return nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m AppModel) View() string {
	return lipgloss.NewStyle().PaddingTop(1).PaddingLeft(2).Render(lipgloss.JoinVertical(
		lipgloss.Top,
		lipgloss.NewStyle().Foreground(lipgloss.Color("30")).Render("Working path:"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("125")).Render("Remote url:"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("74")).Render("Revision:"),
	))
}

func main() {
	p := tea.NewProgram(New(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("error: %v", err)
	}
}
