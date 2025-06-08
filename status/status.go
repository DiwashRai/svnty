package status

import (
	"github.com/DiwashRai/svnty/styles"
	"github.com/DiwashRai/svnty/svn"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	SvnService svn.Service
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *Model) View() string {
	var content string
	for i, section := range m.SvnService.CurrentStatus().Sections {
		content += styles.StatusSectionHeading.Render(svn.SectionTitles[i]) + "\n"
		for _, ps := range section {
			content += styles.InfoStr.Render(string(ps.Status)+" "+ps.Path) + "\n"
		}
		content += "\n"
	}
	return lipgloss.NewStyle().PaddingTop(1).PaddingLeft(2).Render(content)
}
