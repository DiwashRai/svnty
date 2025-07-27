package commit

import (
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/DiwashRai/svnty/styles"
	"github.com/DiwashRai/svnty/svn"
	"github.com/DiwashRai/svnty/tui"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	maxDisplayLength = 72
	commitPanelWidth = 77 // 72 + 4(linenumber gutter) + 1(padding left)
)

var (
	border    = lipgloss.RoundedBorder()
	commitTop = styles.GetBorderTopWithTitle("Commit Message", commitPanelWidth)
)

type CommitMode int

const (
	EditMessageMode CommitMode = iota
	MsgListMode
)

type Model struct {
	SvnService    svn.Service
	Logger        *slog.Logger
	textarea      textarea.Model
	msglist       list.Model
	Mode          CommitMode
	CommitHistory svn.CommitHistory
}

type ItemType struct {
	DisplayText string
	FullMessage string
}

func (i ItemType) FilterValue() string { return i.FullMessage }

func NewItem(msg string) list.Item {
	displayText := strings.ReplaceAll(msg, "\n", " ")
	if len(displayText) > maxDisplayLength {
		displayText = displayText[:maxDisplayLength-3] + "..."
	}
	return ItemType{
		DisplayText: displayText,
		FullMessage: msg,
	}
}

type itemDelegate struct{}

var (
	itemStyle         = styles.BaseStyle.PaddingLeft(4)
	selectedItemStyle = styles.BaseStyle.PaddingLeft(2).
				Foreground(lipgloss.Color(styles.CommitListSelColor))
)

func (d itemDelegate) Height() int  { return 1 }
func (d itemDelegate) Spacing() int { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	li, ok := item.(ItemType)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, li.DisplayText)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	displayStr := fn(str)
	displayStr = displayStr[:len(displayStr)-4] // remove the resetAll terminal code
	fmt.Fprint(w, displayStr)
}

func (m *Model) buildHistoryItems() []list.Item {
	items := []list.Item{}
	for _, msg := range m.CommitHistory.GetHistory() {
		items = append(items, NewItem(msg))
	}
	return items
}

func (m *Model) Init() tea.Cmd {
	ti := textarea.New()
	ti.ShowLineNumbers = true
	ti.Prompt = ""
	ti.FocusedStyle, ti.BlurredStyle = getTextAreaStyle()

	ti.SetWidth(commitPanelWidth)
	ti.SetHeight(8)
	ti.Focus()
	m.textarea = ti

	historyItems := m.buildHistoryItems()
	historyList := list.New(historyItems, itemDelegate{}, 10, 40)
	historyList.SetShowTitle(false)
	historyList.SetShowStatusBar(false)
	historyList.SetShowHelp(false)
	m.msglist = historyList

	return nil
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	m.Logger.Info("CommitModel.Update()")
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		keyStr := msg.String()
		switch keyStr {
		case "esc":
			return tui.StatusMode
		case "ctrl+d": // [d]o Commit
			return m.Submit()
		case "up", "ctrl+k":
			m.msglist.CursorUp()
		case "down", "ctrl+j":
			m.msglist.CursorDown()
		case "tab":
			if selectedItem, ok := m.msglist.SelectedItem().(ItemType); ok {
				m.textarea.SetValue(m.textarea.Value() + selectedItem.FullMessage)
			}
		case "ctrl+r": // [r]eset message
			m.textarea.SetValue("")
		default:
			m.textarea, cmd = m.textarea.Update(msg)
			return cmd
		}
	}

	return nil
}

func (m *Model) View() string {
	commitPanel := styles.BorderStyle.
		BorderTop(false).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(true).
		Render(m.textarea.View())

	return lipgloss.JoinVertical(
		lipgloss.Left,
		commitTop,
		commitPanel,
		m.msglist.View(),
	)
}

func CommitStagedCmd(m *Model) tea.Cmd {
	return func() tea.Msg {
		err := m.SvnService.CommitStaged(m.textarea.Value())
		if err != nil {
			return tui.RenderErrorMsg(err)
		}

		m.CommitHistory.AddMessage(m.textarea.Value())
		m.CommitHistory.SaveToFile()
		m.textarea.SetValue("")
		return tui.CommitSuccessMsg{}
	}
}

func (m *Model) Submit() tea.Cmd {
	return CommitStagedCmd(m)
}

func (m *Model) SaveDraft() {
	draftMsg := m.textarea.Value()
	if len(draftMsg) == 0 {
		return
	}

	m.CommitHistory.AddMessage(draftMsg)
	m.CommitHistory.SaveToFile()
}

func getTextAreaStyle() (textarea.Style, textarea.Style) {
	focused := textarea.Style{
		Base:             styles.BaseStyle,
		CursorLine:       styles.BaseStyle.Background(lipgloss.Color(styles.BgSelected)),
		CursorLineNumber: styles.BaseStyle.Foreground(lipgloss.Color(styles.SpecialColor)),
		EndOfBuffer:      styles.BaseStyle,
		LineNumber:       styles.BaseStyle.Foreground(lipgloss.Color(styles.LineNumberColor)),
		Placeholder:      styles.BaseStyle,
		Prompt:           styles.BaseStyle,
		Text:             styles.BaseStyle,
	}
	// currently unused
	blurred := textarea.Style{
		Base:             styles.BaseStyle,
		CursorLine:       styles.BaseStyle.Background(lipgloss.Color(styles.BgSelected)),
		CursorLineNumber: styles.BaseStyle.Foreground(lipgloss.Color(styles.SpecialColor)),
		EndOfBuffer:      styles.BaseStyle,
		LineNumber:       styles.BaseStyle.Foreground(lipgloss.Color(styles.LineNumberColor)),
		Placeholder:      styles.BaseStyle,
		Prompt:           styles.BaseStyle,
		Text:             styles.BaseStyle,
	}
	return focused, blurred
}
