package status

import (
	"github.com/DiwashRai/svnty/styles"
	"github.com/DiwashRai/svnty/svn"
	"log/slog"
	"reflect"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ElementType int

const (
	HeaderElem ElementType = iota
	PathElem
	DiffElem  // for inline diffs in future
	BlankElem // used for blank lines between sections
)

type Element struct {
	Type        ElementType
	SectionID   svn.SectionIdx
	SectionSize int
	Content     string
	Status      rune
}

type Model struct {
	SvnService svn.Service
	Logger     *slog.Logger
	Panel      []Element
	CursorIdx  int
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	m.Logger.Info("StatusModel.Update()")

	switch msg := msg.(type) {
	case svn.RefreshStatusMsg:
		m.Logger.Info("Refreshing status panel")
		for i, section := range m.SvnService.CurrentStatus().Sections {
			if len(section.Paths) <= 0 {
				continue
			}
			m.Panel = append(m.Panel,
				Element{
					Type:        HeaderElem,
					SectionID:   svn.SectionIdx(i),
					SectionSize: len(m.SvnService.CurrentStatus().Sections[i].Paths),
					Content:     svn.SectionTitles[i],
				},
			)

			if section.Collapsed {
				m.Panel = append(m.Panel, Element{Type: BlankElem})
				continue
			}

			for _, ps := range section.Paths {
				m.Panel = append(m.Panel,
					Element{
						Type:      PathElem,
						SectionID: svn.SectionIdx(i),
						Content:   ps.Path,
						Status:    ps.Status,
					})
			}
			m.Panel = append(m.Panel, Element{Type: BlankElem})
		}
	default:
		m.Logger.Info("Unhandled Msg type.", "type", reflect.TypeOf(msg))
	}
	return nil
}

func (m *Model) Down() {
	if m.CursorIdx < len(m.Panel)-1 {
		m.CursorIdx++
	}
}

func (m *Model) Up() {
	if m.CursorIdx > 0 {
		m.CursorIdx--
	}
}

func (m *Model) View() string {
	var b strings.Builder
	for line, elem := range m.Panel {
		var gutter string
		var headingStyle, runeStyle, textStyle lipgloss.Style
		if line != m.CursorIdx {
			gutter = styles.Gutter
			headingStyle = styles.StatusSectionHeading
			runeStyle = styles.StatusRune
			textStyle = styles.BaseStyle
		} else {
			gutter = styles.SelGutter
			headingStyle = styles.SelStatusSectionHeading
			runeStyle = styles.SelStatusRune
			textStyle = styles.Selected
		}

		b.WriteString(gutter)

		switch elem.Type {
		case HeaderElem:
			b.WriteString(styles.ExpandedHeader)
			b.WriteString(headingStyle.Render(elem.Content))
			b.WriteString(headingStyle.Render(" ("))
			b.WriteString(textStyle.Render(strconv.Itoa(elem.SectionSize)))
			b.WriteString(headingStyle.Render(") "))

		case PathElem:
			b.WriteString(runeStyle.Render(" ", string(elem.Status), " "))
			b.WriteString(textStyle.Render(elem.Content))

		case DiffElem:
		case BlankElem:
		default:
			m.Logger.Error("Invalid element type encountered")
		}
		b.WriteByte('\n')
	}
	return b.String()
}
