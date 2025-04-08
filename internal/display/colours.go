package display

import (
	"github.com/fatih/color"
)

var (
	Green   = color.New(color.FgGreen).SprintFunc()
	Blue    = color.New(color.FgBlue).SprintFunc()
	Yellow  = color.New(color.FgYellow).SprintFunc()
	Red     = color.New(color.FgRed).SprintFunc()
	Magenta = color.New(color.FgMagenta).SprintFunc()
	White   = color.New(color.FgWhite).SprintFunc()

	BoldGreen  = color.New(color.FgGreen, color.Bold).SprintFunc()
	BoldYellow = color.New(color.FgYellow, color.Bold).SprintFunc()
	BoldWhite  = color.New(color.FgWhite, color.Bold).SprintFunc()
	BoldHiCyan = color.New(color.FgHiCyan, color.Bold).SprintFunc()
)

var (
	HeaderColor    = BoldHiCyan
	CreatedColor   = Green
	UpdatedColor   = Yellow
	UnchangedColor = White
	TimeColor      = Magenta
	BoldTitle      = BoldWhite
	BoldCreated    = BoldGreen
	BoldUpdated    = BoldYellow
	BoldUnchanged  = BoldWhite
)

