package overlay

import (
	"fmt"

	"github.com/rivo/tview"
)

type configureOption struct {
	label string
	flag  string
}

var configureOptions = []configureOption{
	{label: "Compact output", flag: "-c"},
	{label: "Use null input", flag: "-n"},
	{label: "Slurp input", flag: "-s"},
	{label: "Raw output", flag: "-r"},
	{label: "Join output", flag: "-j"},
	{label: "ASCII output", flag: "-a"},
	{label: "Read raw strings", flag: "-R"},
	{label: "Monochrome output", flag: "-M"},
	{label: "Force color", flag: "-C"},
	{label: "Sort keys", flag: "-S"},
}

func ConfigureRows(getOptionValue func(flag string) bool) []string {
	rows := make([]string, 0, len(configureOptions))
	for _, option := range configureOptions {
		checkbox := "○"
		if getOptionValue(option.flag) {
			checkbox = "●"
		}

		rows = append(rows, fmt.Sprintf("%s %s (%s)", checkbox, option.label, option.flag))
	}

	return rows
}

func ToggleConfigureOption(index int, getOptionValue func(flag string) bool, setOptionValue func(flag string, value bool)) bool {
	if index < 0 || index >= len(configureOptions) {
		return false
	}

	option := configureOptions[index]
	nextValue := !getOptionValue(option.flag)
	setOptionValue(option.flag, nextValue)

	if option.flag == "-C" && nextValue {
		setOptionValue("-M", false)
	}

	if option.flag == "-M" && nextValue {
		setOptionValue("-C", false)
	}

	return true
}

var configureSize = func() struct {
	Width  int
	Height int
} {
	rows := ConfigureRows(func(flag string) bool {
		return false
	})

	contentWidth := 0
	for _, row := range rows {
		if rowWidth := tview.TaggedStringWidth(row); rowWidth > contentWidth {
			contentWidth = rowWidth
		}
	}

	if contentWidth < len("Configure") {
		contentWidth = len("Configure")
	}

	return struct {
		Width  int
		Height int
	}{
		Width:  contentWidth + 2,
		Height: len(rows) + 3,
	}
}()
