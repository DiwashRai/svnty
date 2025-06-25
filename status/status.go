package status

import (
	"log/slog"
	"reflect"
	"strconv"
	"strings"

	"github.com/DiwashRai/svnty/styles"
	"github.com/DiwashRai/svnty/svn"
	"github.com/DiwashRai/svnty/tui"

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

const (
	HEADER_IDX = -1
	IGNORE_IDX = -2
)

type Element struct {
	Type        ElementType
	SectionID   svn.SectionIdx
	ItemIdx     int
	SectionSize int
	Content     string
	Status      rune
	Expanded    bool
}

type Cursor struct {
	Section svn.SectionIdx
	Item    int
}

type Model struct {
	//Width      int
	Height     int
	YOffset    int
	SvnService svn.Service
	Logger     *slog.Logger
	Panel      []Element
	Cursor     Cursor
	Errs       []string
	Lines      []string
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	m.Logger.Info("StatusModel.Update()")

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Height = msg.Height - 6 // info panel size 4 + 1 padding top
	case tui.FetchStatus:
		return FetchStatusCmd(m.SvnService)
	case tui.RefreshStatus:
		return RefreshStatusCmd(m)
	case tui.RenderError:
		m.Errs = append(m.Errs, msg.Error())
		return nil
	case tea.KeyMsg:
		keyStr := msg.String()
		switch keyStr {
		case "j":
			m.Down()
			return nil
		case "k":
			m.Up()
			return nil
		case "s":
			return m.Stage()
		case "u":
			return m.Unstage()
		case "=":
			m.Diff()
			return nil
		case "enter":
			return m.ToggleExpanded()
		}
	default:
		m.Logger.Info("Unhandled Msg type.", "type", reflect.TypeOf(msg))
	}
	return nil
}

func headerIcon(isSel bool, isExpanded bool) string {
	if isSel {
		if isExpanded {
			return styles.SelExpandedHeader
		} else {
			return styles.SelCollapsedHeader
		}
	} else {
		if isExpanded {
			return styles.ExpandedHeader
		} else {
			return styles.CollapsedHeader
		}
	}
}

func clamp(v, low, high int) int {
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}

func (m *Model) visibleLines(cursorIdx, padding int) (lines []string) {
	if cursorIdx-padding < m.YOffset {
		m.YOffset = max(0, cursorIdx-padding)
	} else if cursorIdx+padding >= m.YOffset+m.Height {
		m.YOffset = min(len(m.Lines)-m.Height, (cursorIdx-(m.Height-1))+2)
	}

	if len(m.Lines) > 0 {
		top := max(0, m.YOffset)
		bottom := clamp(m.YOffset+m.Height, top, len(m.Lines))
		lines = m.Lines[top:bottom]
	}
	return lines
}

func (m *Model) View() string {
	m.Lines = m.Lines[:0]
	m.Lines = append(m.Lines, m.Errs...)

	var cursorIdx int
	for i, elem := range m.Panel {
		var b strings.Builder
		var isSel bool
		var gutter string
		var headingStyle, runeStyle, textStyle lipgloss.Style
		if elem.SectionID == m.Cursor.Section && elem.ItemIdx == m.Cursor.Item {
			isSel = true
			cursorIdx = i
			gutter = styles.SelGutter
			headingStyle = styles.SelStatusSectionHeading
			runeStyle = styles.SelStatusRune
			textStyle = styles.Selected
		} else {
			isSel = false
			gutter = styles.Gutter
			headingStyle = styles.StatusSectionHeading
			runeStyle = styles.StatusRune
			textStyle = styles.BaseStyle
		}

		b.WriteString(gutter)

		switch elem.Type {
		case HeaderElem:
			b.WriteString(headerIcon(isSel, elem.Expanded))
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
		m.Lines = append(m.Lines, b.String())
	}
	return strings.Join(m.visibleLines(cursorIdx, styles.ScrollPadding), "\n")
}

func (m *Model) RefreshStatus() {
	m.Logger.Info("Refreshing status panel")
	m.Panel = m.Panel[:0]
	m.ClampCursor()

	svnStatus := m.SvnService.CurrentStatus()
	for secID, section := range svnStatus.Sections {
		if len(section.Paths) <= 0 {
			continue
		}

		headerExpanded := true
		if !section.Expanded {
			headerExpanded = false
		}
		m.Panel = append(m.Panel,
			Element{
				Type:        HeaderElem,
				SectionID:   svn.SectionIdx(secID),
				ItemIdx:     HEADER_IDX,
				SectionSize: len(section.Paths),
				Content:     svn.SectionTitles[secID],
				Expanded:    headerExpanded,
			},
		)

		if !section.Expanded {
			m.Panel = append(m.Panel, Element{Type: BlankElem, SectionID: -1})
			continue
		}

		for j, ps := range section.Paths {
			m.Panel = append(m.Panel,
				Element{
					Type:      PathElem,
					SectionID: svn.SectionIdx(secID),
					ItemIdx:   j,
					Content:   ps.Path,
					Status:    ps.Status,
				})
		}
		m.Panel = append(m.Panel, Element{Type: BlankElem, ItemIdx: IGNORE_IDX})
	}
	m.Must(m.Cursor.Section == 0 ||
		(m.Cursor.Item >= HEADER_IDX &&
			m.Cursor.Item < len(svnStatus.Sections[m.Cursor.Section].Paths)),
		"Cursor idx out of bounds")
}

func (m *Model) Must(cond bool, msg string) {
	if !cond {
		m.Errs = append(m.Errs, msg)
	}
}

func FetchInfoCmd(s svn.Service) tea.Cmd {
	return func() tea.Msg {
		if err := s.FetchInfo(); err != nil {
			return tui.RenderError(err)
		}
		return tui.RefreshInfo{}
	}
}

func FetchStatusCmd(s svn.Service) tea.Cmd {
	return func() tea.Msg {
		if err := s.FetchStatus(); err != nil {
			return tui.RenderError(err)
		}
		return tui.RefreshStatus{}
	}
}

func RefreshStatusCmd(m *Model) tea.Cmd {
	return func() tea.Msg {
		m.RefreshStatus()
		return nil
	}
}

func StagePathCmd(s svn.Service, path string) tea.Cmd {
	return func() tea.Msg {
		if err := s.StagePath(path); err != nil {
			return tui.RenderError(err)
		}
		return tui.FetchStatus{}
	}
}

func UnstagePathCmd(s svn.Service, path string) tea.Cmd {
	return func() tea.Msg {
		if err := s.UnstagePath(path); err != nil {
			return tui.RenderError(err)
		}
		return tui.FetchStatus{}
	}
}

func ToggleExpandedCmd(s svn.Service, si svn.SectionIdx) tea.Cmd {
	return func() tea.Msg {
		if err := s.ToggleExpanded(si); err != nil {
			return tui.RenderError(err)
		}
		return tui.RefreshStatus{}
	}
}

func (m *Model) Stage() tea.Cmd {
	m.Logger.Info("StatusModel.Stage() called")
	if m.Cursor.Item <= IGNORE_IDX {
		return nil
	}

	if m.Cursor.Item == HEADER_IDX {
		//TODO: stage whole section
		return nil
	}

	ps, err := m.SvnService.GetPathStatus(m.Cursor.Section, m.Cursor.Item)
	if err != nil {
		return nil
	}
	m.Logger.Info("Returning StagePathCmd", "path", ps.Path)
	return StagePathCmd(m.SvnService, ps.Path)
}

func (m *Model) Unstage() tea.Cmd {
	m.Logger.Info("StatusModel.Unstage() called")
	if m.Cursor.Item <= IGNORE_IDX || m.Cursor.Section != svn.SectionStaged {
		return nil
	}

	if m.Cursor.Item == HEADER_IDX {
		//TODO: unstage whole section
	}

	ps, err := m.SvnService.GetPathStatus(svn.SectionStaged, m.Cursor.Item)
	if err != nil {
		return nil
	}
	return UnstagePathCmd(m.SvnService, ps.Path)
}

func (m *Model) Diff() tea.Cmd {
	m.Logger.Info("StatusModel.Diff() called")
	ps, err := m.SvnService.GetPathStatus(svn.SectionStaged, m.Cursor.Item)
	if err != nil {
		return nil
	}
	diff, err := m.SvnService.GetDiff(ps.Path)
	m.Logger.Info(diff)
	return nil
}

func (m *Model) ToggleExpanded() tea.Cmd {
	if m.Cursor.Item != HEADER_IDX {
		return nil
	}
	return ToggleExpandedCmd(m.SvnService, m.Cursor.Section)
}

func (m *Model) Down() bool {
	svnStatus := m.SvnService.CurrentStatus()

	if svnStatus.Sections[m.Cursor.Section].Expanded &&
		m.Cursor.Item < len(svnStatus.Sections[m.Cursor.Section].Paths)-1 {
		m.Cursor.Item++
		return true
	}
	for i := m.Cursor.Section + 1; i < svn.NumSections; i++ {
		if len(svnStatus.Sections[i].Paths) == 0 {
			continue
		}
		m.Cursor.Section = i
		m.Cursor.Item = HEADER_IDX
		return true
	}
	return false
}

func (m *Model) Up() bool {
	svnStatus := m.SvnService.CurrentStatus()

	if svnStatus.Sections[m.Cursor.Section].Expanded &&
		m.Cursor.Item >= 0 && svnStatus.Len(m.Cursor.Section) > 0 {
		m.Cursor.Item--
		return true
	}
	for i := m.Cursor.Section - 1; i >= 0; i-- {
		currSecSize := len(svnStatus.Sections[i].Paths)
		if currSecSize == 0 {
			continue
		}
		m.Cursor.Section = i
		if svnStatus.Sections[m.Cursor.Section].Expanded {
			m.Cursor.Item = currSecSize - 1
		} else {
			m.Cursor.Item = HEADER_IDX
		}
		return true
	}
	return false
}

func (m *Model) ClampCursor() {
	svnStatus := m.SvnService.CurrentStatus()

	secLen := svnStatus.Len(m.Cursor.Section)
	if m.Cursor.Item >= secLen || secLen <= 0 {
		if !m.Up() && !m.Down() {
			m.Cursor.Section = 0
			m.Cursor.Item = HEADER_IDX
		}
	}
}
