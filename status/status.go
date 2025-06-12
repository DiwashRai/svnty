package status

import (
	"github.com/DiwashRai/svnty/styles"
	"github.com/DiwashRai/svnty/svn"
	"log/slog"
	"reflect"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type ElementType int

const (
	HeaderElem ElementType = iota
	FileElem
	DiffElem  // for inline diffs in future
	BlankElem // used for blank lines between sections
)

type Element struct {
	Type      ElementType
	SectionID svn.Section
	Content   string
}

type Model struct {
	SvnService svn.Service
	Logger     *slog.Logger
	Panel      []Element
	CursorIdx  uint64
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	m.Logger.Info("StatusModel.Update()")

	switch msg := msg.(type) {
	case svn.RefreshStatusMsg:
		m.Logger.Info("Refreshing status panel")
	default:
		m.Logger.Info("Unhandled Msg type.", "type", reflect.TypeOf(msg))
	}
	return nil
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
			b.WriteString(styles.SelStatusSectionHeading.Render(svn.SectionTitles[i]))
			b.WriteString(styles.SelStatusSectionHeading.Render("("))
			b.WriteString(strconv.Itoa(len(section)))
			b.WriteString(styles.SelStatusSectionHeading.Render(")"))
		} else {
			b.WriteString(styles.Gutter)
			b.WriteString(styles.Comment.Render("⯆ "))
			b.WriteString(styles.StatusSectionHeading.Render(svn.SectionTitles[i]))
			b.WriteString(styles.StatusSectionHeading.Render("("))
			b.WriteString(strconv.Itoa(len(section)))
			b.WriteString(styles.StatusSectionHeading.Render(")"))
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
