package styles

import "github.com/charmbracelet/lipgloss"

const (
	// color codes from kanagwa.nvim
	fujiWhite   = "#DCD7BA" // kanagawa: foreground color
	oldWhite    = "#C8C093" // kanagawa: foreground dim
	fujiGray    = "#727169" // kanagawa: comment color
	waveRed     = "#E46876" // kanagawa: 'return' red
	springGreen = "#98BB6C" // kanagawa: string green
	sakuraPink  = "#D27E99" // kanagawa: number pink
	oniViolet   = "#957FB8" // kanagawa: keyword purple
	sumiInk3    = "#1F1F28" // kanagawa: bg color

	// semantic color assignments
	bgColor      = sumiInk3
	fgColor      = fujiWhite
	fgDimColor   = oldWhite
	CommentColor = fujiGray
	numColor     = sakuraPink
	keywordColor = oniViolet
	specialColor = waveRed
)

var (
	BaseStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(bgColor))

	InfoHeading = BaseStyle.
			Bold(true).
			Foreground(lipgloss.Color(keywordColor))

	Fg = BaseStyle.
		Foreground(lipgloss.Color(fgColor))

	FgDim = BaseStyle.
		Foreground(lipgloss.Color(fgDimColor))

	Comment = BaseStyle.
		Foreground(lipgloss.Color(CommentColor))

	Number = BaseStyle.
		Foreground(lipgloss.Color(numColor))

	Selected = BaseStyle.
			Foreground(lipgloss.Color(springGreen))

	StatusSectionHeading = BaseStyle.
				Bold(true).
				Foreground(lipgloss.Color(specialColor))
)
