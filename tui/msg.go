package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
)

type StatusModeMsg struct{}
type CommitModeMsg struct{}

type FetchInfoMsg struct{}
type FetchStatusMsg struct{}
type FetchHeadRevisionMsg struct{}

type RefreshInfoMsg struct{}
type RefreshStatusPanelMsg struct{}

type RenderErrorMsg error
type CommitSuccessMsg struct{}
type UpdateSuccessMsg struct{}
type QuitMsg struct{}

type HeadRevisionTickMsg struct{}

func StatusMode() tea.Msg {
	return StatusModeMsg{}
}
func CommitMode() tea.Msg {
	return CommitModeMsg{}
}

func FetchInfo() tea.Msg {
	return FetchInfoMsg{}
}
func FetchStatus() tea.Msg {
	return FetchStatusMsg{}
}

func RefreshInfo() tea.Msg {
	return RefreshInfoMsg{}
}
func RefreshStatusPanel() tea.Msg {
	return RefreshStatusPanelMsg{}
}

func Quit() tea.Msg {
	return QuitMsg{}
}

func FetchHeadRevision() tea.Msg {
	return FetchHeadRevisionMsg{}
}

func HeadRevisionTicker(sec int) tea.Cmd {
	return tea.Tick(time.Duration(sec)*time.Second, func(t time.Time) tea.Msg {
		return HeadRevisionTickMsg{}
	})
}
