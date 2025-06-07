package logging

import (
	"log/slog"
	"os"
)

func New(logPath string) (*slog.Logger, func(), error) {
	if logPath == "" {
		return slog.New(slog.DiscardHandler), func() {}, nil
	}

	logFile, err := os.OpenFile(
		logPath,
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0666,
	)

	if err != nil {
		return nil, nil, err
	}

	var handler slog.Handler
	handler = slog.NewTextHandler(logFile, &slog.HandlerOptions{
		AddSource: true,
	})

	logger := slog.New(handler)
	closeFile := func() {
		_ = logFile.Close()
	}
	return logger, closeFile, nil
}
