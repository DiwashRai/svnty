package app

import (
	"github.com/DiwashRai/svnty/info"
	"github.com/DiwashRai/svnty/svn"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	SvnService svn.Service
	InfoModel  info.Model
}

func New(svc svn.Service) Model {
	return Model{
		SvnService: svc,
		InfoModel:  info.Model{SvnService: svc},
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
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	return m.InfoModel.View()
}
