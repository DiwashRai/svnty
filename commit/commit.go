package commit

import (
	"log/slog"

	"github.com/DiwashRai/svnty/styles"
	"github.com/DiwashRai/svnty/svn"
	"github.com/DiwashRai/svnty/tui"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	border           = lipgloss.RoundedBorder()
	commitPanelWidth = 77 // 72 + 1(eol) + 4(linenumber gutter)
	commitTop        = styles.GetBorderTopWithTitle("Commit Message", commitPanelWidth)
)

type CommitMode int

const (
	EditMessageMode CommitMode = iota
	MsgListMode
)

type Model struct {
	SvnService svn.Service
	Logger     *slog.Logger
	textarea   textarea.Model
	msglist    list.Model
	Mode       CommitMode
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
		case "tab":
			return m.Submit()
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

	return lipgloss.JoinVertical(lipgloss.Left, commitTop, commitPanel)
}

func CommitStagedCmd(m *Model) tea.Cmd {
	return func() tea.Msg {
		err := m.SvnService.CommitStaged(m.textarea.Value())
		if err != nil {
			return tui.RenderErrorMsg(err)
		}
		m.textarea.SetValue("")
		return tui.CommitSuccessMsg{}
	}
}

func (m *Model) Submit() tea.Cmd {
	return CommitStagedCmd(m)
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
