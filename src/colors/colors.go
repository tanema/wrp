package colors

import (
	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
)

var (
	// Red is the color red
	Red = color.New(color.FgRed).SprintFunc()
	// Yellow is the color yellow
	Yellow = color.New(color.FgYellow).SprintFunc()
	// Blue is the color blue
	Blue = color.New(color.FgBlue).SprintFunc()
	// Green is the color Green
	Green = color.New(color.FgGreen).SprintFunc()
	// ColorStdOut is a wrapped std out that allows colors
	ColorStdOut = colorable.NewColorableStdout()
	// ColorStdErr is a wrapped std err that allows colors
	ColorStdErr = colorable.NewColorableStderr()
)
