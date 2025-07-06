package app

import (
	"log/slog"
	"reflect"

	"github.com/DiwashRai/svnty/commit"
	"github.com/DiwashRai/svnty/info"
	"github.com/DiwashRai/svnty/status"
	"github.com/DiwashRai/svnty/styles"
	"github.com/DiwashRai/svnty/svn"
	"github.com/DiwashRai/svnty/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AppMode int

const (
	StatusMode AppMode = iota
	CommitMode
)

type Model struct {
	SvnService  svn.Service
	Logger      *slog.Logger
	InfoModel   info.Model
	StatusModel status.Model
	CommitModel commit.Model
	Mode        AppMode
	width       int
	height      int
}

func New(svc svn.Service, logger *slog.Logger) Model {
	model := Model{
		SvnService: svc,
		Logger:     logger,
		InfoModel:  info.Model{SvnService: svc},
		StatusModel: status.Model{
			SvnService: svc,
			Logger:     logger,
			Cursor:     status.Cursor{ElemType: status.HeaderElem},
		},
		CommitModel: commit.Model{
			SvnService: svc,
			Logger:     logger,
		},
		Mode: StatusMode,
	}

	model.CommitModel.Init()
	return model
}

func (m *Model) Init() tea.Cmd {
	m.Logger.Info("App.Init()")
	m.StatusModel.Init()
	m.SvnService.Init()
	return tea.Batch(
		status.FetchInfoCmd(m.SvnService),
		status.FetchStatusCmd(m.SvnService),
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
		m.StatusModel.Update(msg)
		return m, nil
	case tui.StatusModeMsg:
		m.Mode = StatusMode
		return m, nil
	case tui.CommitModeMsg:
		m.Mode = CommitMode
		return m, nil
	case tui.CommitSuccessMsg:
		return m, tea.Batch(tui.FetchStatus, tui.StatusMode)
	case tui.FetchStatusMsg:
		cmd = m.StatusModel.Update(msg)
		return m, cmd
	case tui.RefreshStatusPanelMsg:
		cmd = m.StatusModel.Update(msg)
		return m, cmd
	case tui.RenderErrorMsg:
		cmd = m.StatusModel.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		keyStr := msg.String()
		m.Logger.Info(keyStr)
		switch keyStr {
		case "ctrl+c":
			return m, tea.Quit
		default:
			switch m.Mode {
			case StatusMode:
				cmd = m.StatusModel.Update(msg)
				return m, cmd
			case CommitMode:
				cmd = m.CommitModel.Update(msg)
				return m, cmd
			}
		}
	default:
		m.Logger.Info("Unhandled Msg type.", "type", reflect.TypeOf(msg))
	}
	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	var content string
	switch m.Mode {
	case StatusMode:
		content = tui.JoinVerticalStyled(
			lipgloss.Left,
			styles.BaseStyle,
			m.InfoModel.View(),
			m.StatusModel.View(),
		)
	case CommitMode:
		/*
			content = tui.JoinVerticalStyled(
				lipgloss.Left,
				styles.BaseStyle,
				m.InfoModel.View(),
				m.CommitModel.View(),
			)
		*/
		content = tui.JoinVerticalStyled(
			lipgloss.Left,
			styles.BaseStyle,
			styles.BaseStyle.Render(m.CommitModel.View()),
		)
	}
	m.Logger.Info("App.View()")
	return styles.BaseStyle.
		PaddingLeft(1).
		PaddingTop(1).
		Width(m.width).
		Height(m.height).
		Render(content)
}
