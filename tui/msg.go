package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type StatusModeMsg struct{}
type CommitModeMsg struct{}

type FetchInfoMsg struct{}
type FetchStatusMsg struct{}

type RefreshInfoMsg struct{}
type RefreshStatusPanelMsg struct{}

type RenderErrorMsg error

func StatusMode() tea.Msg {
	return StatusModeMsg{}
}
func CommitMode() tea.Msg {
	return CommitModeMsg{}
}
