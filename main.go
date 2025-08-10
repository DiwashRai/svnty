package main

import (
	"flag"
	"fmt"
	"github.com/DiwashRai/svnty/app"
	"github.com/DiwashRai/svnty/logging"
	"github.com/DiwashRai/svnty/svn"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	workingCopyPath := flag.String("path", "C:/Code/GitHub/textual-test/", "path of svn repo")
	//workingCopyPath := flag.String("path", ".", "path of svn repo")
	//workingCopyPath := flag.String("path", "C:/Projects/svn-checkouts/textual-test", "path of svn repo")
	useMock := flag.Bool("mock", false, "use mocked SVN data")
	logPath := flag.String("log", "", "write logs to this file")
	flag.Parse()

	rootLogger, closeLogFile, err := logging.New(*logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open log: %v\n", err)
		os.Exit(1)
	}
	defer closeLogFile()

	var svc svn.Service
	if *useMock {
		var mockSvc svn.MockService
		svc = &mockSvc
	} else {
		realSvc := svn.RealService{
			WorkingCopyPath: *workingCopyPath,
			Logger:          rootLogger,
		}
		svc = &realSvc
	}

	model := app.New(svc, rootLogger)

	p := tea.NewProgram(&model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("error: %v", err)
		os.Exit(1)
	}
}
