package info

import (
	"github.com/DiwashRai/svnty/styles"
	"github.com/DiwashRai/svnty/svn"
	"github.com/charmbracelet/lipgloss"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	SvnService svn.Service
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
	info := m.SvnService.CurrentInfo()
	wp, url, rev := info.WorkingPath, info.RemoteURL, strconv.FormatUint(uint64(info.Revision), 10)

	return lipgloss.NewStyle().PaddingTop(1).PaddingLeft(2).Render(lipgloss.JoinVertical(
		lipgloss.Top,
		styles.InfoHeading.Render(" Working path: ")+" "+styles.InfoStr.Render(wp),
		styles.InfoHeading.Render(" Remote URL:   ")+" "+styles.InfoStr.Render(url),
		styles.InfoHeading.Render(" Revision:     ")+" "+styles.Number.Render(rev),
	))
}
