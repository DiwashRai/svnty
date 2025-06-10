package status

import (
	"github.com/DiwashRai/svnty/styles"
	"github.com/DiwashRai/svnty/svn"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	SvnService svn.Service
	CursorIdx  int
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *Model) View() string {
	var content string
	var lineNum uint64
	for i, section := range m.SvnService.CurrentStatus().Sections {
		lnStr := strconv.FormatUint(lineNum, 10)
		content += lnStr + " "
		content += styles.Comment.Render("⯆ ")
		content += styles.StatusSectionHeading.Render(svn.SectionTitles[i]) + "\n"
		lineNum++

		for _, ps := range section {
			var line string
			lnStr := strconv.FormatUint(lineNum, 10)
			line += lnStr + "  "
			if lineNum == 3 {
				line += styles.Selected.Render("> " + string(ps.Status) + " " + ps.Path)
			} else {
				line += styles.Fg.Render("  " + string(ps.Status) + " " + ps.Path)
			}
			content += line + "\n"
			lineNum++
		}
		content += "\n"
	}
	return styles.BaseStyle.
		PaddingTop(1).
		Render(content)
}
