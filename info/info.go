package info

import (
	"github.com/DiwashRai/svnty/styles"
	"github.com/DiwashRai/svnty/svn"
	"github.com/DiwashRai/svnty/utils"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

	return utils.JoinVerticalStyled(
		lipgloss.Left,
		styles.BaseStyle,
		styles.InfoHeading.Render("Working path: ")+styles.Fg.Render(wp),
		styles.InfoHeading.Render("Remote URL:   ")+styles.Fg.Render(url),
		styles.InfoHeading.Render("Revision:     ")+styles.Number.Render(rev),
	)
}
