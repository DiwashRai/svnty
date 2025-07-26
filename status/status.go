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

type Expanded struct {
	section [svn.NumSections]bool
	path    map[string]bool
}

func (e *Expanded) Init() {
	e.path = make(map[string]bool)
}

func (e *Expanded) Path(p string) bool {
	return e.path != nil && e.path[p]
}

func (e *Expanded) SetPath(p string, b bool) {
	e.path[p] = b
}

func (e *Expanded) TogglePath(p string) {
	val, ok := e.path[p]
	if !ok {
		e.path[p] = true
	}
	e.path[p] = !val
}

func (e *Expanded) Section(si svn.SectionIdx) bool {
	return e.section[si]
}

func (e *Expanded) SetSection(si svn.SectionIdx, b bool) {
	e.section[si] = b
}

func (e *Expanded) ToggleSection(si svn.SectionIdx) {
	e.section[si] = !e.section[si]
}

type Model struct {
	Width      int
	Height     int
	YOffset    int
	SvnService svn.Service
	Logger     *slog.Logger
	Panel      []Element
	Cursor     Cursor
	Errs       []string
	Lines      []string
	Expanded   Expanded
}

func (m *Model) Init() tea.Cmd {
	m.Logger.Info("StatusModel.Init() called")

	for i := svn.SectionIdx(0); i < svn.NumSections; i++ {
		m.Expanded.SetSection(i, true)
	}
	m.Expanded.SetSection(svn.SectionUnversioned, false)
	m.Expanded.Init()

	return nil
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	m.Logger.Info("StatusModel.Update()")

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height - 6 // info panel size 4 + 1 padding top
	case tui.FetchStatusMsg:
		return FetchStatusCmd(m.SvnService)
	case tui.RefreshStatusPanelMsg:
		return RefreshStatusPanelCmd(m)
	case tui.RenderErrorMsg:
		m.Errs = append(m.Errs, msg.Error())
		return nil
	case tea.KeyMsg:
		keyStr := msg.String()
		switch keyStr {
		case "c":
			return tui.CommitMode
		case "j":
			m.Down()
			return nil
		case "k":
			m.Up()
			return nil
		case "q":
			return tui.Quit
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

func (m *Model) truncateIfNeeded(content string) string {
	// padding on left is 1
	availableWidth := m.Width - (styles.GutterLen + 1)
	if availableWidth <= 0 {
		return content
	}

	if len(content) <= availableWidth {
		return content
	}

	// Reserve 3 chars for "..."
	if availableWidth <= 3 {
		return "..."
	}

	return content[:availableWidth-3] + "..."
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
			truncatedContent := m.truncateIfNeeded(elem.Content)
			switch {
			case strings.HasPrefix(elem.Content, "@@"):
				b.WriteString(diffHeaderStyle.Render(truncatedContent))
			case strings.HasPrefix(elem.Content, "+"):
				b.WriteString(addedStyle.Render(truncatedContent))
			case strings.HasPrefix(elem.Content, "-"):
				b.WriteString(removedStyle.Render(truncatedContent))
			default:
				b.WriteString(textStyle.Render(truncatedContent))
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

		sectionExpanded := m.Expanded.Section(svn.SectionIdx(secID))
		m.Panel = append(m.Panel,
			Element{
				Type:      HeaderElem,
				SectionID: svn.SectionIdx(secID),
				PathIdx:   0,
				DiffLine:  0,
				Size:      len(section.Paths),
				Content:   svn.SectionTitles[secID],
				Expanded:  sectionExpanded,
			},
		)

		if !sectionExpanded {
			m.Panel = append(m.Panel, Element{Type: BlankElem})
			continue
		}

		for pathIdx, ps := range section.Paths {
			pathExpanded := m.Expanded.Path(ps.Path)
			m.Panel = append(m.Panel,
				Element{
					Type:      PathElem,
					SectionID: svn.SectionIdx(secID),
					PathIdx:   pathIdx,
					DiffLine:  0,
					Content:   ps.Path,
					Status:    ps.Status,
					Expanded:  pathExpanded,
				})

			if !pathExpanded {
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
			return tui.RenderErrorMsg(err)
		}
		return tui.RefreshInfoMsg{}
	}
}

func FetchStatusCmd(s svn.Service) tea.Cmd {
	return func() tea.Msg {
		if err := s.FetchStatus(); err != nil {
			return tui.RenderErrorMsg(err)
		}
		return tui.RefreshStatusPanelMsg{}
	}
}

func RefreshStatusPanelCmd(m *Model) tea.Cmd {
	return func() tea.Msg {
		m.RefreshStatusPanel()
		return nil
	}
}

func StagePathCmd(m *Model, path string) tea.Cmd {
	return func() tea.Msg {
		if err := m.SvnService.StagePath(path); err != nil {
			return tui.RenderErrorMsg(err)
		}
		m.Expanded.SetPath(path, false)
		return tui.FetchStatusMsg{}
	}
}

func UnstagePathCmd(m *Model, path string) tea.Cmd {
	return func() tea.Msg {
		if err := m.SvnService.UnstagePath(path); err != nil {
			return tui.RenderErrorMsg(err)
		}
		m.Expanded.SetPath(path, false)
		return tui.FetchStatusMsg{}
	}
}

func ToggleSectionExpandCmd(m *Model, si svn.SectionIdx) tea.Cmd {
	return func() tea.Msg {
		m.Expanded.ToggleSection(si)
		return tui.RefreshStatusPanelMsg{}
	}
}

func ToggleDiffExpandCmd(m *Model) tea.Cmd {
	return func() tea.Msg {
		ps, err := m.SvnService.GetPathStatus(m.Cursor.Section, m.Cursor.PathIdx)
		if err != nil {
			return tui.RenderErrorMsg(err)
		}

		// case: expanded -> collapsed
		if m.Expanded.Path(ps.Path) {
			m.Expanded.TogglePath(ps.Path)
			m.Cursor.Set(PathElem, m.Cursor.Section, m.Cursor.PathIdx, 0)
			return tui.RefreshStatusPanelMsg{}
		}

		// case: collapsed -> expanded
		if ps.Status != 'M' && ps.Status != 'A' {
			return nil
		}
		if err = m.SvnService.FetchDiff(ps.Path); err != nil {
			return tui.RenderErrorMsg(err)
		}
		m.Expanded.TogglePath(ps.Path)
		return tui.RefreshStatusPanelMsg{}
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
		return StagePathCmd(m, ps.Path)
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
		return UnstagePathCmd(m, ps.Path)
	}

	return nil
}

func (m *Model) Diff() tea.Cmd {
	m.Logger.Info("StatusModel.Diff() called")
	if m.Cursor.ElemType != PathElem && m.Cursor.ElemType != DiffElem {
		return nil
	}
	return ToggleDiffExpandCmd(m)
}

func (m *Model) ToggleSectionExpand() tea.Cmd {
	if m.Cursor.ElemType != HeaderElem {
		return nil
	}
	return ToggleSectionExpandCmd(m, m.Cursor.Section)
}

func (m *Model) nextSectionHeader() bool {
	if next, ok := m.SvnService.CurrentStatus().NextNonEmptySection(m.Cursor.Section); ok {
		m.Cursor.Set(HeaderElem, next, 0, 0)
		return true
	}
	return false
}

func (m *Model) DownFromHeader() bool {
	if m.Cursor.ElemType != HeaderElem {
		return false
	}

	rs := m.SvnService.CurrentStatus()
	// Section is expanded and has entries. Go to first path
	if m.Expanded.Section(m.Cursor.Section) && rs.Len(m.Cursor.Section) > 0 {
		m.Cursor.Set(PathElem, m.Cursor.Section, 0, 0)
		return true
	}

	// Current section end reached or empty. Go to next section
	return m.nextSectionHeader()
}

func (m *Model) DownFromPath() bool {
	if m.Cursor.ElemType != PathElem {
		return false
	}

	ps, err := m.SvnService.GetPathStatus(m.Cursor.Section, m.Cursor.PathIdx)
	if err != nil {
		return false
	}
	// Path expanded so navigate to first line of diff
	if m.Expanded.Path(ps.Path) {
		m.Cursor.Set(DiffElem, m.Cursor.Section, m.Cursor.PathIdx, 0)
		return true
	}
	// Path not expanded so just go to next path
	if m.Cursor.PathIdx < m.SvnService.CurrentStatus().Len(m.Cursor.Section)-1 {
		m.Cursor.Set(PathElem, m.Cursor.Section, m.Cursor.PathIdx+1, 0)
		return true
	}
	// Last path of current section reached. Go to next section
	return m.nextSectionHeader()
}

func (m *Model) DownFromDiff() bool {
	if m.Cursor.ElemType != DiffElem {
		return false
	}

	ps, err := m.SvnService.GetPathStatus(m.Cursor.Section, m.Cursor.PathIdx)
	if err != nil {
		return false
	}
	// Still more lines in current diff
	diffLines := m.SvnService.GetDiff(ps.Path)
	if m.Cursor.DiffLine < len(diffLines)-1 {
		m.Cursor.Set(DiffElem, m.Cursor.Section, m.Cursor.PathIdx, m.Cursor.DiffLine+1)
		return true
	}
	// No more lines in diff. Navigate to next path if not at final path
	rs := m.SvnService.CurrentStatus()
	if m.Cursor.PathIdx < rs.Len(m.Cursor.Section)-1 {
		m.Cursor.Set(PathElem, m.Cursor.Section, m.Cursor.PathIdx+1, 0)
		return true
	}

	return m.nextSectionHeader()
}

func (m *Model) Down() bool {
	switch m.Cursor.ElemType {
	case HeaderElem:
		return m.DownFromHeader()
	case PathElem:
		return m.DownFromPath()
	case DiffElem:
		return m.DownFromDiff()
	default:
		m.Errs = append(m.Errs, "Invalid Cursor ElementType encountered in Down()")
	}
	return false
}

func (m *Model) upFromHeader() bool {
	if m.Cursor.ElemType != HeaderElem {
		return false
	}

	rs := m.SvnService.CurrentStatus()
	prevSec, ok := rs.PrevNonEmptySection(m.Cursor.Section)
	// We are at header of the upper most non empty section. Nothing to move up to
	if !ok {
		return false
	}

	// PrevSec not expanded so navigate straight to header
	if !m.Expanded.Section(prevSec) {
		m.Cursor.Set(HeaderElem, prevSec, 0, 0)
		return true
	}

	prevSecSize := rs.Len(prevSec)
	prevSecLastPath, err := m.SvnService.GetPathStatus(prevSec, prevSecSize-1)
	if err != nil {
		return false
	}

	// PrevPath is collapsed so move up to path
	if !m.Expanded.Path(prevSecLastPath.Path) {
		m.Cursor.Set(PathElem, prevSec, prevSecSize-1, 0)
		return true
	}

	// Path is expanded so move up to last diff line
	diffLines := m.SvnService.GetDiff(prevSecLastPath.Path)
	m.Cursor.Set(DiffElem, prevSec, prevSecSize-1, len(diffLines)-1)
	return true
}

func (m *Model) upFromPath() bool {
	if m.Cursor.ElemType != PathElem {
		return false
	}

	prevPath, err := m.SvnService.GetPathStatus(m.Cursor.Section, m.Cursor.PathIdx-1)
	if err == nil {
		if m.Expanded.Path(prevPath.Path) {
			diffLines := m.SvnService.GetDiff(prevPath.Path)
			m.Cursor.Set(DiffElem, m.Cursor.Section, m.Cursor.PathIdx-1, max(0, len(diffLines)-1))
		} else {
			m.Cursor.Set(PathElem, m.Cursor.Section, m.Cursor.PathIdx-1, 0)
		}
		return true
	}

	m.Cursor.Set(HeaderElem, m.Cursor.Section, 0, 0)
	return true
}

func (m *Model) upFromDiff() bool {
	if m.Cursor.ElemType != DiffElem {
		return false
	}

	if m.Cursor.DiffLine > 0 {
		m.Cursor.Set(DiffElem, m.Cursor.Section, m.Cursor.PathIdx, m.Cursor.DiffLine-1)
	} else if m.Cursor.DiffLine == 0 {
		m.Cursor.Set(PathElem, m.Cursor.Section, m.Cursor.PathIdx, 0)
	}

	return true
}

func (m *Model) Up() bool {
	switch m.Cursor.ElemType {
	case HeaderElem:
		return m.upFromHeader()
	case PathElem:
		return m.upFromPath()
	case DiffElem:
		return m.upFromDiff()
	default:
		m.Errs = append(m.Errs, "Invalid Cursor ElementType encountered in Up()")
	}
	return false
}

func (m *Model) ClampCursor() {
	rs := m.SvnService.CurrentStatus()

	secLen := rs.Len(m.Cursor.Section)
	if m.Cursor.PathIdx >= secLen || secLen <= 0 {
		if secLen == 0 {
			m.Cursor.Set(HeaderElem, m.Cursor.Section, 0, 0)
		}
		if !m.Up() && !m.Down() {
			m.Cursor.Section = 0
			m.Cursor.PathIdx = 0
			m.Cursor.DiffLine = 0
		}
	}
}
