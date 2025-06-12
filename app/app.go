package app

import (
	"github.com/DiwashRai/svnty/info"
	"github.com/DiwashRai/svnty/status"
	"github.com/DiwashRai/svnty/styles"
	"github.com/DiwashRai/svnty/svn"
	"github.com/DiwashRai/svnty/utils"
	"log/slog"
	"reflect"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	SvnService  svn.Service
	Logger      *slog.Logger
	InfoModel   info.Model
	StatusModel status.Model
	width       int
	height      int
}

func New(svc svn.Service, logger *slog.Logger) Model {
	return Model{
		SvnService:  svc,
		Logger:      logger,
		InfoModel:   info.Model{SvnService: svc},
		StatusModel: status.Model{SvnService: svc, Logger: logger},
	}
}

func (m *Model) Init() tea.Cmd {
	m.Logger.Info("App.Init()")
	return tea.Batch(
		svn.FetchInfoCmd(m.SvnService),
		svn.FetchStatusCmd(m.SvnService),
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	m.Logger.Info("App.Update()")

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case svn.RefreshStatusMsg:
		cmd = m.StatusModel.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		keyStr := msg.String()
		m.Logger.Info(keyStr)
		switch keyStr {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "j":
			m.StatusModel.Down()
			return m, nil
		case "k":
			m.StatusModel.Up()
			return m, nil
		}
	default:
		m.Logger.Info("Unhandled Msg type.", "type", reflect.TypeOf(msg))
	}
	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	content := utils.JoinVerticalStyled(
		lipgloss.Left,
		styles.BaseStyle,
		m.InfoModel.View(),
		m.StatusModel.View(),
	)
	return styles.BaseStyle.
		PaddingLeft(1).
		PaddingTop(1).
		Width(m.width).
		Height(m.height).
		Render(content)
}
