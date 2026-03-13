package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/require"

	"codeberg.org/gpanders/ijq/internal/options"
)

const (
	testScreenWidth  = 120
	testScreenHeight = 40

	testStartupTimeout = 3 * time.Second
	testActionTimeout  = 2 * time.Second
	testPollInterval   = 10 * time.Millisecond
	testRedrawDelay    = 100 * time.Millisecond

	testJQCommand = "./testdata/catok"
)

type testApp struct {
	t           *testing.T
	app         *tview.Application
	screen      tcell.SimulationScreen
	runErr      chan error
	historyPath string
}

func newTestApp(t *testing.T, input string, historyEntries []string) *testApp {
	t.Helper()

	if runtime.GOOS == "windows" {
		t.Skip("ui tests rely on the shell-based testdata/catok helper")
	}

	cfg := DefaultConfig()
	cfg.JQCommand = testJQCommand
	cfg.HistoryFile = ""

	historyPath := ""
	if historyEntries != nil {
		historyPath = filepath.Join(t.TempDir(), "history")
		contents := ""
		if len(historyEntries) > 0 {
			contents = strings.Join(historyEntries, "\n") + "\n"
		}
		require.NoError(t, os.WriteFile(historyPath, []byte(contents), 0o644))
		cfg.HistoryFile = options.HistoryFile(historyPath)
	}

	doc := Document{
		input:  input,
		filter: ".",
		options: options.Options{
			HistoryFile: cfg.HistoryFile,
			JQCommand:   cfg.JQCommand,
		},
		config: cfg,
	}

	app := createApp(doc)
	screen := tcell.NewSimulationScreen("")
	require.NoError(t, screen.Init())
	app.SetScreen(screen)
	screen.SetSize(testScreenWidth, testScreenHeight)

	ta := &testApp{
		t:           t,
		app:         app,
		screen:      screen,
		runErr:      make(chan error, 1),
		historyPath: historyPath,
	}

	go func() {
		ta.runErr <- app.Run()
	}()

	ta.waitForText("Input (Top)", testStartupTimeout)
	ta.waitForText("Output (Top)", testStartupTimeout)
	ta.waitForText("Filter", testStartupTimeout)
	ta.waitForText("Error", testStartupTimeout)
	ta.waitForText("menu", testStartupTimeout)

	t.Cleanup(ta.stop)

	return ta
}

func (ta *testApp) stop() {
	if ta.runErr == nil {
		return
	}

	ta.app.Stop()

	select {
	case err := <-ta.runErr:
		require.NoError(ta.t, err)
	case <-time.After(5 * time.Second):
		ta.t.Fatalf("timed out waiting for app to stop\n%s", ta.screenContent())
	}

	ta.runErr = nil
}

func (ta *testApp) postKey(key tcell.Key, mod tcell.ModMask) {
	ta.t.Helper()
	ta.app.QueueEvent(tcell.NewEventKey(key, ' ', mod))
	if key != tcell.KeyCtrlC {
		time.Sleep(testRedrawDelay)
	}
}

func (ta *testApp) postRune(r rune) {
	ta.t.Helper()
	ta.app.QueueEvent(tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone))
	time.Sleep(testRedrawDelay)
}

func (ta *testApp) postAltRune(r rune) {
	ta.t.Helper()
	ta.app.QueueEvent(tcell.NewEventKey(tcell.KeyRune, r, tcell.ModAlt))
	time.Sleep(testRedrawDelay)
}

func (ta *testApp) postRunes(text string) {
	ta.t.Helper()
	for _, r := range text {
		ta.postRune(r)
	}
}

func (ta *testApp) rows() []string {
	rows := []string(nil)
	ta.app.QueueUpdate(func() {
		cells, width, height := ta.screen.GetContents()
		rows = make([]string, height)

		for y := range height {
			runes := make([]rune, width)
			for x := range width {
				cell := cells[y*width+x]
				runes[x] = ' '
				if len(cell.Runes) > 0 {
					runes[x] = cell.Runes[0]
				}
			}
			rows[y] = string(runes)
		}
	})

	return rows
}

func (ta *testApp) row(y int) string {
	ta.t.Helper()
	rows := ta.rows()
	require.GreaterOrEqual(ta.t, y, 0)
	require.Less(ta.t, y, len(rows))
	return rows[y]
}

func (ta *testApp) findRowOf(text string) int {
	for i, row := range ta.rows() {
		if strings.Contains(row, text) {
			return i
		}
	}

	return -1
}

func (ta *testApp) findOnScreen(text string) bool {
	return ta.findRowOf(text) >= 0
}

func (ta *testApp) screenContent() string {
	rows := ta.rows()
	trimmed := make([]string, len(rows))
	for i, row := range rows {
		trimmed[i] = strings.TrimRight(row, " ")
	}
	return strings.Join(trimmed, "\n")
}

func (ta *testApp) waitFor(cond func() bool, description string, timeout time.Duration) {
	ta.t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(testPollInterval)
	}

	ta.t.Fatalf("timed out waiting for %s\n%s", description, ta.screenContent())
}

func (ta *testApp) waitForText(text string, timeout time.Duration) {
	ta.t.Helper()
	ta.waitFor(func() bool {
		return ta.findOnScreen(text)
	}, fmt.Sprintf("%q to appear on screen", text), timeout)
}

func (ta *testApp) waitForNoText(text string, timeout time.Duration) {
	ta.t.Helper()
	ta.waitFor(func() bool {
		return !ta.findOnScreen(text)
	}, fmt.Sprintf("%q to disappear from screen", text), timeout)
}

func (ta *testApp) waitForInputFieldFocus(timeout time.Duration) {
	ta.t.Helper()
	ta.waitFor(func() bool {
		_, ok := ta.app.GetFocus().(*tview.InputField)
		return ok
	}, "filter input to have focus", timeout)
}

func (ta *testApp) waitForTextViewFocus(timeout time.Duration) {
	ta.t.Helper()
	ta.waitFor(func() bool {
		_, ok := ta.app.GetFocus().(*tview.TextView)
		return ok
	}, "text view to have focus", timeout)
}

func (ta *testApp) requireText(text string) {
	ta.t.Helper()
	require.Truef(ta.t, ta.findOnScreen(text), "expected to find %q on screen\n%s", text, ta.screenContent())
}

func (ta *testApp) requireNoText(text string) {
	ta.t.Helper()
	require.Falsef(ta.t, ta.findOnScreen(text), "expected not to find %q on screen\n%s", text, ta.screenContent())
}

func (ta *testApp) openMenu() {
	ta.t.Helper()
	ta.postKey(tcell.KeyCtrlUnderscore, tcell.ModNone)
	ta.waitForText("Menu", testActionTimeout)
	ta.waitForText("Configure", testActionTimeout)
}

func (ta *testApp) selectMenuItem(index int) {
	ta.t.Helper()
	for range index {
		ta.postKey(tcell.KeyDown, tcell.ModNone)
	}
	ta.postKey(tcell.KeyEnter, tcell.ModNone)
}

func generateLargeInput(lines int) string {
	rows := make([]string, lines)
	for i := range lines {
		rows[i] = fmt.Sprintf(`{"line":%d,"value":"row-%03d"}`, i, i)
	}
	return strings.Join(rows, "\n")
}

func TestUILayout(t *testing.T) {
	ta := newTestApp(t, `{"key":"value"}`, nil)

	inputRow := ta.findRowOf("Input (Top)")
	outputRow := ta.findRowOf("Output (Top)")
	filterRow := ta.findRowOf("Filter")
	errorRow := ta.findRowOf("Error")
	helpRow := ta.findRowOf("menu")

	require.NotEqual(t, -1, inputRow, ta.screenContent())
	require.NotEqual(t, -1, outputRow, ta.screenContent())
	require.Equal(t, inputRow, outputRow, ta.screenContent())
	require.NotEqual(t, -1, filterRow, ta.screenContent())
	require.NotEqual(t, -1, errorRow, ta.screenContent())
	require.NotEqual(t, -1, helpRow, ta.screenContent())

	topRow := ta.row(inputRow)
	inputCol := strings.Index(topRow, "Input (Top)")
	outputCol := strings.Index(topRow, "Output (Top)")
	require.NotEqual(t, -1, inputCol, ta.screenContent())
	require.NotEqual(t, -1, outputCol, ta.screenContent())
	require.Less(t, inputCol, outputCol, ta.screenContent())

	require.Less(t, inputRow, filterRow, ta.screenContent())
	require.Less(t, filterRow, errorRow, ta.screenContent())
	require.Less(t, errorRow, helpRow, ta.screenContent())

	ta.requireText("quit")
	ta.requireText("quit and write output")
}

func TestUIFilterInput(t *testing.T) {
	ta := newTestApp(t, `{"key":"value"}`, nil)

	ta.requireText(".")

	ta.postRunes("foo")
	ta.waitForText(".foo", testActionTimeout)

	ta.postRunes(".bar")
	ta.waitForText(".foo.bar", testActionTimeout)

	for range 3 {
		ta.postKey(tcell.KeyBackspace2, tcell.ModNone)
	}
	ta.waitForText(".foo.", testActionTimeout)
	ta.requireNoText(".foo.bar")
}

func TestUIFocusMovement(t *testing.T) {
	ta := newTestApp(t, generateLargeInput(100), nil)

	expectedFilter := "."
	ta.waitForInputFieldFocus(testActionTimeout)
	ta.postRune('x')
	expectedFilter += "x"
	ta.waitForText(expectedFilter, testActionTimeout)

	ta.postKey(tcell.KeyUp, tcell.ModShift)
	ta.waitForTextViewFocus(testActionTimeout)
	ta.postRune('G')
	ta.waitForText("Input (Bot)", testActionTimeout)

	ta.postKey(tcell.KeyTab, tcell.ModNone)
	ta.waitForTextViewFocus(testActionTimeout)
	ta.postRune('G')
	ta.waitForText("Output (Bot)", testActionTimeout)

	ta.postKey(tcell.KeyTab, tcell.ModNone)
	ta.waitForInputFieldFocus(testActionTimeout)
	ta.postRune('y')
	expectedFilter += "y"
	ta.waitForText(expectedFilter, testActionTimeout)

	ta.postKey(tcell.KeyRight, tcell.ModShift)
	ta.waitForTextViewFocus(testActionTimeout)
	ta.postRune('b')
	ta.waitForNoText("Output (Bot)", testActionTimeout)

	ta.postKey(tcell.KeyDown, tcell.ModShift)
	ta.waitForInputFieldFocus(testActionTimeout)
	ta.postRune('z')
	expectedFilter += "z"
	ta.waitForText(expectedFilter, testActionTimeout)

	ta.postKey(tcell.KeyLeft, tcell.ModShift)
	ta.waitForTextViewFocus(testActionTimeout)
	ta.postRune('b')
	ta.waitForNoText("Input (Bot)", testActionTimeout)

	ta.postKey(tcell.KeyBacktab, tcell.ModNone)
	ta.waitForInputFieldFocus(testActionTimeout)
	ta.postRune('w')
	expectedFilter += "w"
	ta.waitForText(expectedFilter, testActionTimeout)
}

func TestUIScrollKeybindingsAndIndicator(t *testing.T) {
	ta := newTestApp(t, generateLargeInput(100), nil)

	ta.postKey(tcell.KeyUp, tcell.ModShift)
	ta.waitForTextViewFocus(testActionTimeout)
	ta.waitForText("Input (Top)", testActionTimeout)

	ta.postKey(tcell.KeyCtrlN, tcell.ModNone)
	ta.waitForText("Input (1%)", testActionTimeout)

	ta.postKey(tcell.KeyCtrlP, tcell.ModNone)
	ta.waitForText("Input (Top)", testActionTimeout)

	ta.postKey(tcell.KeyCtrlD, tcell.ModNone)
	ta.waitForText("Input (15%)", testActionTimeout)

	ta.postKey(tcell.KeyCtrlU, tcell.ModNone)
	ta.waitForText("Input (Top)", testActionTimeout)

	ta.postRune('G')
	ta.waitForText("Input (Bot)", testActionTimeout)

	ta.postRune('b')
	ta.waitForNoText("Input (Bot)", testActionTimeout)
	ta.requireNoText("Input (Top)")

	ta.postRune('f')
	ta.waitForText("Input (Bot)", testActionTimeout)

	ta.postKey(tcell.KeyRight, tcell.ModShift)
	ta.waitForTextViewFocus(testActionTimeout)
	ta.waitForText("Output (Top)", testActionTimeout)

	ta.postKey(tcell.KeyCtrlD, tcell.ModNone)
	ta.waitForText("Output (15%)", testActionTimeout)

	ta.postRune('G')
	ta.waitForText("Output (Bot)", testActionTimeout)

	ta.postRune('b')
	ta.waitForNoText("Output (Bot)", testActionTimeout)
	ta.requireNoText("Output (Top)")

	ta.postRune('f')
	ta.waitForText("Output (Bot)", testActionTimeout)
}

func TestUIOverlayMenuToggle(t *testing.T) {
	ta := newTestApp(t, `{"key":"value"}`, nil)

	ta.openMenu()
	ta.requireText("Configure")
	ta.requireText("Save current filter to history")
	ta.requireText("Open focused pane in editor")
	ta.requireText("Manage history")
	ta.requireText("Keybindings")
	ta.requireText("Cheat sheet")
	ta.requireText("close")
	ta.requireText("select")

	ta.postKey(tcell.KeyCtrlUnderscore, tcell.ModNone)
	ta.waitForNoText("Configure", testActionTimeout)
	ta.waitForText("Input (Top)", testActionTimeout)
}

func TestUIOverlayMenuConfigure(t *testing.T) {
	ta := newTestApp(t, `{"key":"value"}`, nil)

	rows := []string{
		"Compact output (-c)",
		"Use null input (-n)",
		"Slurp input (-s)",
		"Raw output (-r)",
		"Join output (-j)",
		"ASCII output (-a)",
		"Read raw strings (-R)",
		"Monochrome output (-M)",
		"Force color (-C)",
		"Sort keys (-S)",
		"Hide input (left) viewing pane (-hide-input-pane)",
	}

	ta.openMenu()
	ta.selectMenuItem(0)
	ta.waitForText("Space/Enter", testActionTimeout)
	ta.waitForText("toggle", testActionTimeout)

	for _, row := range rows {
		ta.requireText("○ " + row)
	}

	for i, row := range rows {
		ta.postRune(' ')
		ta.waitForText("● "+row, testActionTimeout)
		ta.waitForNoText("○ "+row, testActionTimeout)

		ta.postRune(' ')
		ta.waitForText("○ "+row, testActionTimeout)
		ta.waitForNoText("● "+row, testActionTimeout)

		if i < len(rows)-1 {
			ta.postKey(tcell.KeyDown, tcell.ModNone)
		}
	}
}

func TestUIOverlayMenuManageHistory(t *testing.T) {
	ta := newTestApp(t, `{"key":"value"}`, []string{".foo", ".bar", ".baz"})

	ta.openMenu()
	ta.selectMenuItem(2)
	ta.waitForText("showing 3 of 3 entries", testActionTimeout)
	ta.requireText(".foo")
	ta.requireText(".bar")
	ta.requireText(".baz")

	ta.postRune('X')
	ta.waitForText("Delete the following entry from history?", testActionTimeout)
	ta.requireText(".foo")
	ta.requireText("Yes")
	ta.requireText("No")

	ta.postKey(tcell.KeyEnter, tcell.ModNone)
	ta.waitForText("showing 2 of 2 entries", testActionTimeout)
	ta.waitForNoText(".foo", testActionTimeout)
	ta.requireText(".bar")
	ta.requireText(".baz")

	contents, err := os.ReadFile(ta.historyPath)
	require.NoError(t, err)
	require.Equal(t, ".bar\n.baz\n", string(contents))

	ta.postRune('/')
	ta.waitForText("Filter:", testActionTimeout)
	ta.postRunes("baz")
	ta.waitForText("showing 1 of 2 entries", testActionTimeout)
	ta.requireText(".baz")
	ta.requireNoText(".bar")
}

func TestUIOverlayMenuCheatSheet(t *testing.T) {
	ta := newTestApp(t, `{"key":"value"}`, nil)

	ta.openMenu()
	ta.selectMenuItem(5)
	ta.waitForText("jq cheat sheet", testActionTimeout)
	ta.requireText("Basics")
	ta.requireText("identity (return input)")
	ta.requireText("Selection / filtering")
	ta.requireText("Object / array building")
	ta.requireText("Common transforms")
	ta.requireText("Strings and formatting")
	ta.requireText("Tip")
	ta.requireText("test incrementally")
	ta.requireText("close")
}

func TestUIOverlayMenuKeybindings(t *testing.T) {
	ta := newTestApp(t, `{"key":"value"}`, nil)

	ta.openMenu()
	ta.selectMenuItem(4)
	ta.waitForText("Keybindings", testActionTimeout)
	ta.requireText("submit-filter")
	ta.requireText("Enter")
	ta.requireText("move-down")
	ta.requireText("toggle-menu")
	ta.requireText("next-focus")
	ta.requireText("Tab")
	ta.requireText("textview-end")
	ta.requireText("quit")
	ta.requireText("Ctrl-C")
	ta.requireText("close")
}

func TestUIOpenEditorNoopPreservesContent(t *testing.T) {
	t.Setenv("VISUAL", "true")

	ta := newTestApp(t, `{"key":"value"}`, nil)

	// Type a filter so we have known content in the filter input
	ta.postRunes("key")
	ta.waitForText(".key", testActionTimeout)

	// Open editor (Alt+E) — "true" exits 0 without modifying the file,
	// so the original content should be preserved after scissor stripping.
	ta.postAltRune('e')

	// App should resume and filter should still show the same text
	ta.waitForText(".key", testActionTimeout)
	ta.waitForText("Filter", testActionTimeout)
}

func TestUIOpenEditorFilterWriteBack(t *testing.T) {
	t.Setenv("VISUAL", "./testdata/editor-write")
	t.Setenv("EDITOR_CONTENT", ".new_filter")

	ta := newTestApp(t, `{"key":"value"}`, nil)

	ta.requireText(".")
	ta.waitForInputFieldFocus(testActionTimeout)

	ta.postAltRune('e')
	ta.waitForText(".new_filter", testActionTimeout)
}

func TestUIOpenEditorInputWriteBack(t *testing.T) {
	t.Setenv("VISUAL", "./testdata/editor-write")
	t.Setenv("EDITOR_CONTENT", `{"replaced":true}`)

	ta := newTestApp(t, `{"original":true}`, nil)

	ta.waitForText("original", testActionTimeout)

	// Move focus to input pane
	ta.postKey(tcell.KeyUp, tcell.ModShift)
	ta.waitForTextViewFocus(testActionTimeout)

	ta.postAltRune('e')
	ta.waitForText("replaced", testActionTimeout)
}

func TestUIOpenEditorOutputReadOnly(t *testing.T) {
	t.Setenv("VISUAL", "./testdata/editor-write")
	t.Setenv("EDITOR_CONTENT", "SHOULD_NOT_APPEAR")

	ta := newTestApp(t, `{"key":"value"}`, nil)

	ta.waitForText("key", testActionTimeout)

	// Move focus to output pane
	ta.postKey(tcell.KeyRight, tcell.ModShift)
	ta.waitForTextViewFocus(testActionTimeout)

	ta.postAltRune('e')

	// Wait for the app to resume after the editor exits, then verify
	// the write-back content did not leak into any pane.
	ta.waitForText("Filter", testActionTimeout)
	ta.requireNoText("SHOULD_NOT_APPEAR")
	ta.requireText("key")
}

func TestUIOpenEditorError(t *testing.T) {
	t.Setenv("VISUAL", "false")

	ta := newTestApp(t, `{"key":"value"}`, nil)

	ta.postAltRune('e')
	ta.waitForText("Editor error", testActionTimeout)
}

func TestUIOpenEditorViaMenu(t *testing.T) {
	t.Setenv("VISUAL", "./testdata/editor-write")
	t.Setenv("EDITOR_CONTENT", ".menu_filter")

	ta := newTestApp(t, `{"key":"value"}`, nil)

	ta.waitForInputFieldFocus(testActionTimeout)

	ta.openMenu()
	ta.selectMenuItem(3)

	ta.waitForText(".menu_filter", testActionTimeout)
}
