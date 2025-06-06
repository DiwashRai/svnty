package info

import (
	"github.com/DiwashRai/svnty/svn"
	"github.com/charmbracelet/lipgloss"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	SvnService svn.Service
	RepoInfo   svn.RepoInfo
}

func New(svc svn.Service) Model {
	return Model{SvnService: svc}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *Model) View() string {
	return lipgloss.NewStyle().PaddingTop(1).PaddingLeft(2).Render(lipgloss.JoinVertical(
		lipgloss.Top,
		lipgloss.NewStyle().Foreground(lipgloss.Color("30")).
			Render("Working path: ", m.SvnService.CurrentInfo().WorkingPath),
		lipgloss.NewStyle().Foreground(lipgloss.Color("125")).
			Render("Remote url: ", m.SvnService.CurrentInfo().RemoteURL),
		lipgloss.NewStyle().Foreground(lipgloss.Color("74")).
			Render("Revision: ", strconv.FormatUint(uint64(m.SvnService.CurrentInfo().Revision), 10)),
	))
}
