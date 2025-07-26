package svn

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
)

const (
	MaxHistorySize  = 10
	HistoryFileName = "commit_history.json"
)

type CommitHistory struct {
	Messages    []string `json:"messages"`
	disabled    bool
	historyFile string
	logger      *slog.Logger
}

func NewCommitHistory(logger *slog.Logger) CommitHistory {
	configDir, err := os.UserConfigDir()
	if err != nil {
		logger.Warn("Failed to get user config directory, disabling history", "error", err)
		return CommitHistory{
			Messages: []string{},
			disabled: true,
			logger:   logger,
		}
	}

	svntyDir := filepath.Join(configDir, "svnty")
	err = os.MkdirAll(svntyDir, 0755)
	if err != nil {
		logger.Warn("Failed to create config directory, disabling history", "error", err)
		return CommitHistory{
			Messages: []string{},
			disabled: true,
			logger:   logger,
		}
	}

	historyFile := filepath.Join(svntyDir, HistoryFileName)
	history := CommitHistory{
		Messages:    []string{},
		disabled:    false,
		historyFile: historyFile,
		logger:      logger,
	}

	history.LoadFromFile()
	return history
}

func (ch *CommitHistory) LoadFromFile() {
	if ch.disabled {
		return
	}

	data, err := os.ReadFile(ch.historyFile)
	if err != nil {
		if !os.IsNotExist(err) {
			ch.logger.Warn("Failed to read history file", "error", err)
		}
		return
	}

	err = json.Unmarshal(data, ch)
	if err != nil {
		ch.logger.Warn("Failed to parse history file", "error", err)
	}
}

func (ch *CommitHistory) GetHistory() []string {
	return ch.Messages
}

func (ch *CommitHistory) AddMessage(msg string) {
	ch.Messages = slices.Insert(ch.Messages, 0, msg)
	if len(ch.Messages) > MaxHistorySize {
		ch.Messages = ch.Messages[:MaxHistorySize]
	}
}

func (ch *CommitHistory) SaveToFile() {
	if ch.disabled {
		return
	}

	jsonData, err := json.MarshalIndent(ch, "", "  ")
	if err != nil {
		ch.logger.Warn("Failed to marshal commit history to JSON", "error", err)
		return
	}

	err = os.WriteFile(ch.historyFile, jsonData, 0644)
	if err != nil {
		ch.logger.Warn("Failed to write commit history file", "error", err, "file", ch.historyFile)
	}
}
