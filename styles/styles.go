package styles

import "github.com/charmbracelet/lipgloss"

const (
	// color codes from kanagwa.nvim
	fujiWhite    = "#DCD7BA" // kanagawa: foreground color
	oldWhite     = "#C8C093" // kanagawa: foreground dim
	fujiGray     = "#727169" // kanagawa: comment color
	waveRed      = "#E46876" // kanagawa: 'return' red
	springGreen  = "#98BB6C" // kanagawa: string green
	sakuraPink   = "#D27E99" // kanagawa: number pink
	oniViolet    = "#957FB8" // kanagawa: keyword purple
	waveAqua2    = "#7AA89F" // kanagawa: type color
	sumiInk3     = "#1F1F28" // kanagawa: bg color
	sumiInk4     = "#2A2A37" // kanagawa: gutter color
	surimiOrange = "#FFA066" // kanagawa: const color
	waveBlue1    = "#223249" // kanagawa: visual block
	boatYellow2  = "#C0A36E" // kanagawa: terminal yellow
	autumnGreen  = "#76946A" // kanagawa: vcs added
	autumnRed    = "#C34043" // kanagawa: vcs removed

	// semantic color assignments
	BgColor       = sumiInk3
	BgAltColor    = sumiInk4
	BgSelected    = waveBlue1
	FgColor       = fujiWhite
	FgDimColor    = oldWhite
	CommentColor  = fujiGray
	NumColor      = sakuraPink
	KeywordColor  = oniViolet
	SpecialColor  = surimiOrange
	Special2Color = waveRed

	DiffHeaderColor = waveAqua2
	AddedColor      = autumnGreen
	RemovedColor    = autumnRed

	ScrollPadding = 2
)

var (
	BaseStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(BgColor)).
			Foreground(lipgloss.Color(FgColor))

	GutterStyle = BaseStyle.
			Background(lipgloss.Color(BgAltColor))

	FgDim = BaseStyle.
		Foreground(lipgloss.Color(FgDimColor))

	Comment = BaseStyle.
		Foreground(lipgloss.Color(CommentColor))

	Number = BaseStyle.
		Foreground(lipgloss.Color(NumColor))

	Selected = BaseStyle.
			Background(lipgloss.Color(BgSelected))

	// Info panel
	InfoHeading = BaseStyle.
			Bold(true).
			Foreground(lipgloss.Color(KeywordColor))

	// Status panel
	StatusSectionHeading = BaseStyle.
				Bold(true).
				Foreground(lipgloss.Color(Special2Color))

	SelStatusSectionHeading = StatusSectionHeading.
				Background(lipgloss.Color(BgSelected))

	SelComment = Comment.
			Background(lipgloss.Color(BgSelected))

	SelNumber = Number.
			Background(lipgloss.Color(BgSelected))

	StatusRune = BaseStyle.
			Foreground(lipgloss.Color(SpecialColor))

	SelStatusRune = StatusRune.
			Background(lipgloss.Color(BgSelected))

	// diff colors
	RemovedStyle = BaseStyle.
			Foreground(lipgloss.Color(RemovedColor))

	SelRemovedStyle = RemovedStyle.
			Background(lipgloss.Color(BgSelected))

	AddedStyle = BaseStyle.
			Foreground(lipgloss.Color(AddedColor))

	SelAddedStyle = AddedStyle.
			Background(lipgloss.Color(BgSelected))

	DiffHeaderStyle = BaseStyle.
			Foreground(lipgloss.Color(DiffHeaderColor))

	SelDiffHeaderStyle = DiffHeaderStyle.
				Background(lipgloss.Color(BgSelected))

	// Rendered components
	Gutter    = GutterStyle.Render("    ") + BaseStyle.Render(" ")
	SelGutter = GutterStyle.
			Bold(true).
			Foreground(lipgloss.Color(boatYellow2)).
			Render(" => ") + BaseStyle.Render(" ")

	ExpandedHeader    = Comment.Render("⯆ ")
	SelExpandedHeader = Comment.
				Background(lipgloss.Color(BgSelected)).
				Render("⯆ ")

	CollapsedHeader    = Comment.Render("▶ ")
	SelCollapsedHeader = Comment.
				Background(lipgloss.Color(BgSelected)).
				Render("▶ ")
)
