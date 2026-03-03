package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"codeberg.org/gpanders/ijq/internal/options"
)

func TestUsageHidesLegacyFlags(t *testing.T) {
	opts := options.Options{}

	var out bytes.Buffer
	flagSet, _, _ := newFlagSet("ijq", &opts, &out)
	flagSet.Usage()

	help := out.String()
	assert.Contains(t, help, "-c")
	assert.Contains(t, help, "-f")
	assert.NotContains(t, help, "-H")
	assert.NotContains(t, help, "-jqbin")
	assert.NotContains(t, help, "-hide-input-pane")
}

func TestLegacyFlagsRemainSupported(t *testing.T) {
	opts := options.Options{}

	var out bytes.Buffer
	flagSet, _, _ := newFlagSet("ijq", &opts, &out)
	err := flagSet.Parse([]string{"-H", "", "-jqbin", "custom-jq", "-hide-input-pane"})
	assert.NoError(t, err)

	assert.Equal(t, options.HistoryFile(""), opts.HistoryFile)
	assert.Equal(t, options.JQCommand("custom-jq"), opts.JQCommand)
	assert.Equal(t, options.HideInputPane(true), opts.HideInputPane)
}

func TestLibraryPathsFromConfigAndFlags(t *testing.T) {
	opts := options.Options{
		LibraryPaths: options.LibraryPaths{"/config/modules"},
	}

	var out bytes.Buffer
	flagSet, _, _ := newFlagSet("ijq", &opts, &out)
	err := flagSet.Parse([]string{"-L", "/cli/modules"})
	assert.NoError(t, err)

	assert.Equal(t, options.LibraryPaths{"/config/modules", "/cli/modules"}, opts.LibraryPaths)
}

func TestLegacyFlagsOverrideConfigValues(t *testing.T) {
	opts := options.Options{
		HistoryFile:   options.HistoryFile("/config/history"),
		JQCommand:     options.JQCommand("/usr/bin/jq"),
		HideInputPane: options.HideInputPane(false),
	}

	var out bytes.Buffer
	flagSet, _, _ := newFlagSet("ijq", &opts, &out)
	err := flagSet.Parse([]string{"-H", "", "-jqbin", "custom-jq", "-hide-input-pane"})
	assert.NoError(t, err)

	assert.Equal(t, options.HistoryFile(""), opts.HistoryFile)
	assert.Equal(t, options.JQCommand("custom-jq"), opts.JQCommand)
	assert.Equal(t, options.HideInputPane(true), opts.HideInputPane)
}
