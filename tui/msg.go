package tui

import (
	tea "github.com/charmbracelet/bubbletea/v2"
)

type StatusModeMsg struct{}
type CommitModeMsg struct{}

type FetchInfoMsg struct{}
type FetchStatusMsg struct{}

type RefreshInfoMsg struct{}
type RefreshStatusPanelMsg struct{}

type RenderErrorMsg error
type CommitSuccessMsg struct{}
type QuitMsg struct{}

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
