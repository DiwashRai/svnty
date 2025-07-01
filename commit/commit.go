package commit

import (
	"log/slog"

	"github.com/DiwashRai/svnty/styles"
	"github.com/DiwashRai/svnty/svn"
	"github.com/DiwashRai/svnty/tui"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	SvnService svn.Service
	Logger     *slog.Logger
	textarea   textarea.Model
}

func (m *Model) Init() tea.Cmd {
	ti := textarea.New()
	ti.FocusedStyle, ti.BlurredStyle = getTextAreaStyle()
	ti.SetHeight(8)
	ti.SetWidth(72)

	//ti.Placeholder = "Commit message ..."
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
		default:
			m.textarea, cmd = m.textarea.Update(msg)
			return cmd
		}
	}

	return nil
}

func (m *Model) View() string {
	return m.textarea.View()
}

func getTextAreaStyle() (textarea.Style, textarea.Style) {
	focused := textarea.Style{
		Base:             styles.BaseStyle,
		CursorLine:       styles.BaseStyle.Background(lipgloss.AdaptiveColor{Light: "255", Dark: "0"}),
		CursorLineNumber: styles.BaseStyle.Foreground(lipgloss.AdaptiveColor{Light: "240"}),
		EndOfBuffer:      styles.BaseStyle.Foreground(lipgloss.AdaptiveColor{Light: "254", Dark: "0"}),
		LineNumber:       styles.BaseStyle.Foreground(lipgloss.AdaptiveColor{Light: "249", Dark: "7"}),
		Placeholder:      styles.BaseStyle.Foreground(lipgloss.Color("240")),
		Prompt:           styles.BaseStyle.Foreground(lipgloss.Color("7")),
		Text:             styles.BaseStyle,
	}
	blurred := textarea.Style{
		Base:             styles.BaseStyle,
		CursorLine:       styles.BaseStyle.Foreground(lipgloss.AdaptiveColor{Light: "245", Dark: "7"}),
		CursorLineNumber: styles.BaseStyle.Foreground(lipgloss.AdaptiveColor{Light: "249", Dark: "7"}),
		EndOfBuffer:      styles.BaseStyle.Foreground(lipgloss.AdaptiveColor{Light: "254", Dark: "0"}),
		LineNumber:       styles.BaseStyle.Foreground(lipgloss.AdaptiveColor{Light: "249", Dark: "7"}),
		Placeholder:      styles.BaseStyle.Foreground(lipgloss.Color("240")),
		Prompt:           styles.BaseStyle.Foreground(lipgloss.Color("7")),
		Text:             styles.BaseStyle.Foreground(lipgloss.AdaptiveColor{Light: "245", Dark: "7"}),
	}
	return focused, blurred
}
