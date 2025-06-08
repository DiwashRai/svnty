package styles

import "github.com/charmbracelet/lipgloss"

const (
	fujiWhite   = "#DCD7BA" // kanagawa: foreground color
	waveRed     = "#E46876" // kanagawa: 'return' red
	springGreen = "#98BB6C" // kanagawa: string green
	sakuraPink  = "#D27E99" // kanagawa: number pink
	oniViolet   = "#957FB8" // kanagawa: keyword purple
)

var (
	InfoHeading = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(oniViolet))

	InfoStr = lipgloss.NewStyle().
		Foreground(lipgloss.Color(fujiWhite))

	Number = lipgloss.NewStyle().
		Foreground(lipgloss.Color(sakuraPink))

	StatusSectionHeading = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(waveRed))
)
