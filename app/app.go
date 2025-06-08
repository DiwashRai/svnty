package app

import (
	"github.com/DiwashRai/svnty/info"
	"github.com/DiwashRai/svnty/status"
	"github.com/DiwashRai/svnty/svn"
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	SvnService  svn.Service
	Logger      *slog.Logger
	InfoModel   info.Model
	StatusModel status.Model
}

func New(svc svn.Service, logger *slog.Logger) Model {
	return Model{
		SvnService:  svc,
		Logger:      logger,
		InfoModel:   info.Model{SvnService: svc},
		StatusModel: status.Model{SvnService: svc},
	}
}

func (m *Model) Init() tea.Cmd {
	m.SvnService.FetchInfo()
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
	case tea.KeyMsg:
		keyStr := msg.String()
		m.Logger.Info(keyStr)
		switch keyStr {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Top,
		m.InfoModel.View(),
		m.StatusModel.View(),
	)
}
