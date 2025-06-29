package commit

import (
	"github.com/DiwashRai/svnty/svn"
	"github.com/DiwashRai/svnty/tui"
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	SvnService svn.Service
	Logger     *slog.Logger
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	m.Logger.Info("CommitModel.Update()")

	switch msg := msg.(type) {
	case tea.KeyMsg:
		keyStr := msg.String()
		switch keyStr {
		case "esc":
			return tui.StatusMode
		}
	}

	return nil
}

func (m *Model) View() string {
	return "CommitMode Screen"
}
