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

type Element struct {
	Type      ElementType
	SectionID svn.SectionIdx
	PathIdx   int
	DiffLine  int
	Size      int
	Content   string
	Status    rune
	Expanded  bool
}

type Cursor struct {
	ElemType ElementType
	Section  svn.SectionIdx
	PathIdx  int
	DiffLine int
}

func (c *Cursor) Set(e ElementType, sec svn.SectionIdx, pi, dl int) {
	c.ElemType = e
	c.Section = sec
	c.PathIdx = pi
	c.DiffLine = dl
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
	case tui.RefreshStatusPanel:
		return RefreshStatusPanelCmd(m)
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
			return m.Diff()
		case "enter":
			return m.ToggleSectionExpand()
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
		var addedStyle, removedStyle, diffHeaderStyle lipgloss.Style
		if elem.Type == m.Cursor.ElemType && elem.SectionID == m.Cursor.Section &&
			elem.PathIdx == m.Cursor.PathIdx && elem.DiffLine == m.Cursor.DiffLine {
			isSel = true
			cursorIdx = i
			gutter = styles.SelGutter
			headingStyle = styles.SelStatusSectionHeading
			runeStyle = styles.SelStatusRune
			textStyle = styles.Selected
			addedStyle = styles.SelAddedStyle
			removedStyle = styles.SelRemovedStyle
			diffHeaderStyle = styles.SelDiffHeaderStyle
		} else {
			isSel = false
			gutter = styles.Gutter
			headingStyle = styles.StatusSectionHeading
			runeStyle = styles.StatusRune
			textStyle = styles.BaseStyle
			addedStyle = styles.AddedStyle
			removedStyle = styles.RemovedStyle
			diffHeaderStyle = styles.DiffHeaderStyle
		}

		b.WriteString(gutter)

		switch elem.Type {
		case HeaderElem:
			b.WriteString(headerIcon(isSel, elem.Expanded))
			b.WriteString(headingStyle.Render(elem.Content))
			b.WriteString(headingStyle.Render(" ("))
			b.WriteString(textStyle.Render(strconv.Itoa(elem.Size)))
			b.WriteString(headingStyle.Render(") "))

		case PathElem:
			b.WriteString(runeStyle.Render(" ", string(elem.Status), " "))
			b.WriteString(textStyle.Render(elem.Content))

		case DiffElem:
			switch {
			case strings.HasPrefix(elem.Content, "@@"):
				b.WriteString(diffHeaderStyle.Render(elem.Content))
			case strings.HasPrefix(elem.Content, "+"):
				b.WriteString(addedStyle.Render(elem.Content))
			case strings.HasPrefix(elem.Content, "-"):
				b.WriteString(removedStyle.Render(elem.Content))
			default:
				b.WriteString(textStyle.Render(elem.Content))
			}
		case BlankElem:
		default:
			m.Logger.Error("Invalid element type encountered")
		}
		m.Lines = append(m.Lines, b.String())
	}
	return strings.Join(m.visibleLines(cursorIdx, styles.ScrollPadding), "\n")
}

func (m *Model) RefreshStatusPanel() {
	m.Logger.Info("Refreshing status panel")
	m.Panel = m.Panel[:0]
	m.ClampCursor()

	rs := m.SvnService.CurrentStatus()
	for secID, section := range rs.Sections {
		if len(section.Paths) <= 0 {
			continue
		}

		headerExpanded := true
		if !section.Expanded {
			headerExpanded = false
		}
		m.Panel = append(m.Panel,
			Element{
				Type:      HeaderElem,
				SectionID: svn.SectionIdx(secID),
				PathIdx:   0,
				DiffLine:  0,
				Size:      len(section.Paths),
				Content:   svn.SectionTitles[secID],
				Expanded:  headerExpanded,
			},
		)

		if !section.Expanded {
			m.Panel = append(m.Panel, Element{Type: BlankElem})
			continue
		}

		for pathIdx, ps := range section.Paths {
			m.Panel = append(m.Panel,
				Element{
					Type:      PathElem,
					SectionID: svn.SectionIdx(secID),
					PathIdx:   pathIdx,
					DiffLine:  0,
					Content:   ps.Path,
					Status:    ps.Status,
					Expanded:  ps.Expanded,
				})

			if !ps.Expanded {
				continue
			}

			for lineNum, diffLine := range m.SvnService.GetDiff(ps.Path) {
				m.Panel = append(m.Panel,
					Element{
						Type:      DiffElem,
						SectionID: svn.SectionIdx(secID),
						PathIdx:   pathIdx,
						DiffLine:  lineNum,
						Content:   diffLine,
					})
			}
		}
		m.Panel = append(m.Panel, Element{Type: BlankElem})
	}
	m.Must(m.Cursor.Section == 0 ||
		(m.Cursor.PathIdx >= 0 &&
			m.Cursor.PathIdx < len(rs.Sections[m.Cursor.Section].Paths)),
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
		return tui.RefreshStatusPanel{}
	}
}

func RefreshStatusPanelCmd(m *Model) tea.Cmd {
	return func() tea.Msg {
		m.RefreshStatusPanel()
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

func ToggleSectionExpandCmd(s svn.Service, si svn.SectionIdx) tea.Cmd {
	return func() tea.Msg {
		if err := s.ToggleSectionExpand(si); err != nil {
			return tui.RenderError(err)
		}
		return tui.RefreshStatusPanel{}
	}
}

func ToggleDiffExpandCmd(s svn.Service, si svn.SectionIdx, pathIdx int) tea.Cmd {
	return func() tea.Msg {
		ps, err := s.GetPathStatus(si, pathIdx)
		if err != nil {
			return tui.RenderError(err)
		}

		// case: expanded -> collapsed
		if ps.Expanded {
			if err = s.TogglePathExpand(si, pathIdx); err != nil {
				return tui.RenderError(err)
			}
			return tui.RefreshStatusPanel{}
		}

		// case: collapsed -> expanded
		if ps.Status != 'M' && ps.Status != 'A' {
			return nil
		}
		if err = s.FetchDiff(ps.Path); err != nil {
			return tui.RenderError(err)
		}
		if err = s.TogglePathExpand(si, pathIdx); err != nil {
			return tui.RenderError(err)
		}
		return tui.RefreshStatusPanel{}
	}
}

func (m *Model) Stage() tea.Cmd {
	m.Logger.Info("StatusModel.Stage() called")

	switch m.Cursor.ElemType {
	case HeaderElem:
		//TODO: stage whole section
		return nil
	case PathElem:
		ps, err := m.SvnService.GetPathStatus(m.Cursor.Section, m.Cursor.PathIdx)
		if err != nil {
			return nil
		}
		m.Logger.Info("Returning StagePathCmd", "path", ps.Path)
		return StagePathCmd(m.SvnService, ps.Path)
	}

	return nil
}

func (m *Model) Unstage() tea.Cmd {
	m.Logger.Info("StatusModel.Unstage() called")
	if m.Cursor.Section != svn.SectionStaged {
		return nil
	}

	switch m.Cursor.ElemType {
	case HeaderElem:
		//TODO: stage whole section
		return nil
	case PathElem:
		ps, err := m.SvnService.GetPathStatus(svn.SectionStaged, m.Cursor.PathIdx)
		if err != nil {
			return nil
		}
		return UnstagePathCmd(m.SvnService, ps.Path)
	}

	return nil
}

func (m *Model) Diff() tea.Cmd {
	m.Logger.Info("StatusModel.Diff() called")
	if m.Cursor.ElemType != PathElem {
		return nil
	}
	return ToggleDiffExpandCmd(m.SvnService, m.Cursor.Section, m.Cursor.PathIdx)
}

func (m *Model) ToggleSectionExpand() tea.Cmd {
	if m.Cursor.ElemType != HeaderElem {
		return nil
	}
	return ToggleSectionExpandCmd(m.SvnService, m.Cursor.Section)
}

func (m *Model) nextSectionHeader() bool {
	rs := m.SvnService.CurrentStatus()

	if next, ok := rs.NextNonEmptySection(m.Cursor.Section); ok {
		m.Cursor.Set(HeaderElem, next, 0, 0)
		return true
	}
	return false
}

func (m *Model) nextPath() bool {
	rs := m.SvnService.CurrentStatus()

	switch m.Cursor.ElemType {
	case HeaderElem:
		if rs.Sections[m.Cursor.Section].Expanded && rs.Len(m.Cursor.Section) > 0 {
			m.Cursor.Set(PathElem, m.Cursor.Section, 0, 0)
			return true
		}

	case DiffElem, PathElem:
		if m.Cursor.PathIdx < rs.Len(m.Cursor.Section)-1 {
			m.Cursor.Set(PathElem, m.Cursor.Section, m.Cursor.PathIdx+1, 0)
			return true
		}
	}
	return false
}

func (m *Model) nextDiffLine() bool {
	getPS := func() (svn.PathStatus, bool) {
		ps, err := m.SvnService.GetPathStatus(m.Cursor.Section, m.Cursor.PathIdx)
		if err != nil {
			return svn.PathStatus{}, false
		}
		return ps, true
	}

	switch m.Cursor.ElemType {
	case PathElem:
		ps, ok := getPS()
		if !ok {
			return false
		}
		if ps.Expanded {
			m.Cursor.Set(DiffElem, m.Cursor.Section, m.Cursor.PathIdx, 0)
			return true
		}

	case DiffElem:
		ps, ok := getPS()
		if !ok {
			return false
		}

		diffLines := m.SvnService.GetDiff(ps.Path)
		if m.Cursor.DiffLine < len(diffLines)-1 {
			m.Cursor.Set(DiffElem, m.Cursor.Section, m.Cursor.PathIdx, m.Cursor.DiffLine+1)
			return true
		}
	}
	return false
}

func (m *Model) Down() bool {
	if m.nextDiffLine() || m.nextPath() || m.nextSectionHeader() {
		return true
	}
	return false
}

func (m *Model) prevSectionHeader() bool {
	rs := m.SvnService.CurrentStatus()

	switch m.Cursor.ElemType {
	case HeaderElem:
		if prevSec, ok := rs.PrevNonEmptySection(m.Cursor.Section); ok {
			if !rs.Sections[prevSec].Expanded {
				m.Cursor.Set(HeaderElem, prevSec, 0, 0)
				return true
			}
		}
	case PathElem:
		if m.Cursor.PathIdx == 0 {
			m.Cursor.Set(HeaderElem, m.Cursor.Section, 0, 0)
			return true
		}
	}
	return false
}

func (m *Model) prevPath() bool {
	rs := m.SvnService.CurrentStatus()

	switch m.Cursor.ElemType {
	case HeaderElem:
		if prevSec, ok := rs.PrevNonEmptySection(m.Cursor.Section); ok {
			if rs.Sections[prevSec].Expanded {
				m.Cursor.Set(PathElem, prevSec, rs.Len(prevSec)-1, 0)
				return true
			}
		}

	case PathElem:
		if m.Cursor.PathIdx > 0 {
			m.Cursor.Set(PathElem, m.Cursor.Section, m.Cursor.PathIdx-1, 0)
			return true
		}

	case DiffElem:
		if m.Cursor.DiffLine == 0 {
			m.Cursor.Set(PathElem, m.Cursor.Section, m.Cursor.PathIdx, 0)
			return true
		}
	}
	return false
}

func (m *Model) prevDiffLine() bool {
	rs := m.SvnService.CurrentStatus()

	switch m.Cursor.ElemType {
	case HeaderElem:
		prevSec, ok := rs.PrevNonEmptySection(m.Cursor.Section)
		if !ok || !rs.Sections[prevSec].Expanded {
			return false
		}
		prevSecSize := rs.Len(prevSec)
		prevSecLastPath, err := m.SvnService.GetPathStatus(prevSec, prevSecSize-1)
		if err != nil {
			return false
		}
		if prevSecLastPath.Expanded {
			diffLines := m.SvnService.GetDiff(prevSecLastPath.Path)
			m.Cursor.Set(DiffElem, prevSec, prevSecSize-1, len(diffLines)-1)
			return true
		}

	case PathElem:
		prevPath, err := m.SvnService.GetPathStatus(m.Cursor.Section, m.Cursor.PathIdx-1)
		if err != nil {
			return false
		}
		if prevPath.Expanded {
			diffLines := m.SvnService.GetDiff(prevPath.Path)
			m.Cursor.Set(DiffElem, m.Cursor.Section, m.Cursor.PathIdx-1, max(0, len(diffLines)-1))
			return true
		}

	case DiffElem:
		if m.Cursor.DiffLine > 0 {
			m.Cursor.Set(DiffElem, m.Cursor.Section, m.Cursor.PathIdx, m.Cursor.DiffLine-1)
			return true
		}
	}
	return false
}

func (m *Model) Up() bool {
	if m.prevDiffLine() || m.prevPath() || m.prevSectionHeader() {
		return true
	}
	return false
}

func (m *Model) ClampCursor() {
	rs := m.SvnService.CurrentStatus()

	secLen := rs.Len(m.Cursor.Section)
	if m.Cursor.PathIdx >= secLen || secLen <= 0 {
		if !m.Up() && !m.Down() {
			m.Cursor.Section = 0
			m.Cursor.PathIdx = 0
			m.Cursor.DiffLine = 0
		}
	}
}
