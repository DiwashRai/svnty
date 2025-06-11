package status

import (
	"github.com/DiwashRai/svnty/styles"
	"github.com/DiwashRai/svnty/svn"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	SvnService svn.Service
	CursorIdx  uint64
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *Model) Down() {
	m.CursorIdx++
}

func (m *Model) Up() {
	if m.CursorIdx > 0 {
		m.CursorIdx--
	}
}

func (m *Model) View() string {
	var b strings.Builder
	var lineNum uint64
	for i, section := range m.SvnService.CurrentStatus().Sections {
		if m.CursorIdx == lineNum {
			b.WriteString(styles.SelGutter)
			b.WriteString(styles.SelComment.Render("⯆ "))
			b.WriteString(styles.SelStatusSectionHeading.
				Render(svn.SectionTitles[i], " "))
		} else {
			b.WriteString(styles.Gutter)
			b.WriteString(styles.Comment.Render("⯆ "))
			b.WriteString(styles.StatusSectionHeading.
				Render(svn.SectionTitles[i]))
		}
		b.WriteByte('\n')
		lineNum++

		for _, ps := range section {
			if m.CursorIdx == lineNum {
				b.WriteString(styles.SelGutter)
				b.WriteString(styles.SelStatusRune.Render(" ", string(ps.Status), " "))
				b.WriteString(styles.Selected.Render(ps.Path, " "))
			} else {
				b.WriteString(styles.Gutter)
				b.WriteString(styles.StatusRune.Render(" ", string(ps.Status), " "))
				b.WriteString(styles.BaseStyle.Render(ps.Path))
			}
			b.WriteByte('\n')
			lineNum++
		}

		if m.CursorIdx == lineNum {
			b.WriteString(styles.SelGutter)
		} else {
			b.WriteString(styles.Gutter)
		}
		b.WriteByte('\n')
		lineNum++
	}
	return b.String()
}
