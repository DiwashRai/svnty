package info

import (
	"github.com/DiwashRai/svnty/styles"
	"github.com/DiwashRai/svnty/svn"
	"github.com/DiwashRai/svnty/tui"
	"strconv"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
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

	return tui.JoinVerticalStyled(
		lipgloss.Left,
		styles.BaseStyle,
		styles.Gutter+styles.InfoHeading.Render("Working path: ")+styles.BaseStyle.Render(wp),
		styles.Gutter+styles.InfoHeading.Render("Remote URL:   ")+styles.BaseStyle.Render(url),
		styles.Gutter+styles.InfoHeading.Render("Revision:     ")+styles.Number.Render(rev),
		styles.Gutter,
	)
}
