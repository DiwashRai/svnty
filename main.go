package main

import (
	"flag"
	"fmt"
	"github.com/DiwashRai/svnty/app"
	"github.com/DiwashRai/svnty/svn"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	useMock := flag.Bool("mock", false, "use mocked SVN data")
	flag.Parse()

	var svc svn.Service
	if *useMock {
		var mockSvc svn.MockService
		svc = &mockSvc
	} else {
		var realSvc svn.RealService
		svc = &realSvc
	}

	model := app.New(svc)

	p := tea.NewProgram(&model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("error: %v", err)
	}
}
