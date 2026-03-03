// Copyright (C) 2021 Gregory Anders <greg@gpanders.com>
// Copyright (C) 2021 Herby Gillot <herby.gillot@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"

	"codeberg.org/gpanders/ijq/internal/options"
	"codeberg.org/gpanders/ijq/internal/overlay"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/term"
)

// Special characters that, if present in a JSON key, need to be quoted in the
// jq filter
const specialChars string = ".-:$/"

const alphabet string = "abcdefghijklmnopqrstuvwxyz"

var Version string

type Document struct {
	input   string
	filter  string
	options options.Options
	config  Config
	ctx     context.Context
}

func (d Document) WithFilter(filter string) Document {
	d.filter = filter
	d.ctx = context.Background()
	return d
}

func (d *Document) ReadFrom(r io.Reader) (n int64, err error) {
	var buf bytes.Buffer
	n, err = buf.ReadFrom(r)
	d.input = buf.String()
	return n, err
}

func (d Document) WriteTo(w io.Writer) (n int64, err error) {
	opts := d.options
	if p, ok := w.(*pane); ok {
		// Writer is a pane, so set options accordingly
		opts.ForceColor = true
		opts.Monochrome = false
		opts.CompactOutput = false
		opts.RawOutput = false
		w = tview.ANSIWriter(p)

		// Mark the pane as dirty so the text view is cleared before
		// new output is written.
		p.dirty = true
		defer func() {
			if p.dirty && err == nil {
				// If there was no error and the pane is still marked as dirty that
				// means jq didn't emit any output, so we need to clear the pane
				// manually
				p.tv.Clear()
			}
		}()
	}

	args := append(opts.ToSlice(), d.filter)
	cmd := exec.CommandContext(d.ctx, string(d.options.JQCommand), args...)

	var b bytes.Buffer
	cmd.Stdin = strings.NewReader(d.input)
	cmd.Stdout = w
	cmd.Stderr = &b

	if err := cmd.Run(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			exiterr.Stderr = b.Bytes()
		}
		return 0, err
	}

	return 0, nil
}

type pane struct {
	tv    *tview.TextView
	dirty bool
}

func (pane *pane) Write(p []byte) (n int, err error) {
	if pane.dirty {
		pane.tv.Clear()
		pane.dirty = false
	}

	return pane.tv.Write(p)
}

func newFlagSet(name string, options *options.Options, output io.Writer) (*flag.FlagSet, *string, *bool) {
	flagSet := flag.NewFlagSet(name, flag.ExitOnError)
	flagSet.SetOutput(output)
	flagSet.Usage = func() {
		fmt.Fprintf(output, "ijq - interactive jq\n\n")
		fmt.Fprintf(output, "Usage: ijq [-cnsrRMSV] [-f file] [filter] [files ...]\n\n")
		fmt.Fprintf(output, "Options:\n")

		flagSet.VisitAll(func(f *flag.Flag) {
			// Do not show deprecated flags
			switch f.Name {
			case "H", "hide-input-pane", "jqbin":
				return
			}

			name, usage := flag.UnquoteUsage(f)
			if name == "" {
				fmt.Fprintf(output, "  -%s    \t%s\n", f.Name, usage)
			} else {
				fmt.Fprintf(output, "  -%s %s    \t%s\n", f.Name, name, usage)
			}
		})
	}

	flagSet.Var(&options.CompactOutput, options.CompactOutput.Flag(), "compact instead of pretty-printed output")
	flagSet.Var(&options.NullInput, options.NullInput.Flag(), "use `null` as the single input value")
	flagSet.Var(&options.Slurp, options.Slurp.Flag(), "read all inputs into an array and use it as the single input value")
	flagSet.Var(&options.RawOutput, options.RawOutput.Flag(), "output strings without escapes and quotes")
	flagSet.Var(&options.JoinOutput, options.JoinOutput.Flag(), "implies -r and output without newline after each output")
	flagSet.Var(&options.ASCIIOutput, options.ASCIIOutput.Flag(), "output strings by only ASCII characters using escape sequences")
	flagSet.Var(&options.RawInput, options.RawInput.Flag(), "read each line as string instead of JSON")
	flagSet.Var(&options.ForceColor, options.ForceColor.Flag(), "colorize JSON output")
	flagSet.Var(&options.Monochrome, options.Monochrome.Flag(), "disable colored output")
	flagSet.Var(&options.SortKeys, options.SortKeys.Flag(), "sort keys of each object on output")
	flagSet.Var(&options.LibraryPaths, options.LibraryPaths.Flag(), "search modules from the `dir`ectory")

	// Legacy options kept for backward compatibility.
	flagSet.Var(&options.HideInputPane, options.HideInputPane.Flag(), "hide input (left) viewing pane")
	flagSet.Var(&options.JQCommand, options.JQCommand.Flag(), "name of or path to jq binary to use")
	flagSet.Var(&options.HistoryFile, options.HistoryFile.Flag(), "set path to history file. Set to '' to disable history.")

	filterFile := flagSet.String("f", "", "load the filter from a `file`")
	version := flagSet.Bool("V", false, "print version and exit")

	return flagSet, filterFile, version
}

func parseArgs(options *options.Options) (string, []string) {
	flagSet, filterFile, version := newFlagSet("ijq", options, os.Stderr)
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		log.Fatalln(err)
	}

	if *version {
		fmt.Println("ijq " + Version)
		os.Exit(0)
	}

	filter := "."
	args := flagSet.Args()

	stdinIsTty := term.IsTerminal(int(os.Stdin.Fd()))

	if *filterFile != "" {
		contents, err := os.ReadFile(*filterFile)
		if err != nil {
			log.Fatalln(err)
		}

		filter = string(contents)
	} else if len(args) > 1 || (len(args) > 0 && (!stdinIsTty || bool(options.NullInput))) {
		filter = args[0]
		args = args[1:]
	} else if len(args) == 0 && stdinIsTty && !bool(options.NullInput) {
		flagSet.Usage()
		os.Exit(1)
	}

	return filter, args
}

func scrollHalfPage(tv *tview.TextView, up bool) {
	_, _, _, height := tv.GetInnerRect()
	row, col := tv.GetScrollOffset()
	if up {
		tv.ScrollTo(row-height/2, col)
	} else {
		tv.ScrollTo(row+height/2, col)
	}
}

func scrollHorizontally(tv *tview.TextView, end bool) {
	if end {
		text := tv.GetText(true)
		_, _, width, height := tv.GetInnerRect()
		row, _ := tv.GetScrollOffset()
		maxLen := 0
		for i, line := range strings.Split(text, "\n") {
			if i < row {
				continue
			}

			if i > row+height {
				break
			}

			if length := len(line); length > maxLen {
				maxLen = length
			}
		}

		if maxLen > width {
			tv.ScrollTo(row, maxLen-width)
		}
	} else {
		row, _ := tv.GetScrollOffset()
		tv.ScrollTo(row, 0)
	}
}

func updateScrollIndicator(name string, lineCount int, tv *tview.TextView) {
	row, _ := tv.GetScrollOffset()
	if row <= 0 {
		tv.SetTitle(fmt.Sprintf("%s (Top)", name))
		return
	}

	_, _, _, height := tv.GetInnerRect()
	if row+height >= lineCount {
		tv.SetTitle(fmt.Sprintf("%s (Bot)", name))
		return
	}

	percent := row * 100 / lineCount
	tv.SetTitle(fmt.Sprintf("%s (%d%%)", name, percent))
}

func centerPrimitive(width int, height int, primitive tview.Primitive) tview.Primitive {
	return tview.NewGrid().
		SetRows(0, height, 0).
		SetColumns(0, width, 0).
		AddItem(primitive, 1, 1, 1, 1, 0, 0, true)
}

func normalizeOverlayEvent(event *tcell.EventKey, keymap Keymap) *tcell.EventKey {
	if keymap.MoveDown.Matches(event) {
		return tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone)
	}

	if keymap.MoveUp.Matches(event) {
		return tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModNone)
	}

	if keymap.PageDown.Matches(event) || keymap.HalfPageDown.Matches(event) {
		return tcell.NewEventKey(tcell.KeyPgDn, ' ', tcell.ModNone)
	}

	if keymap.HalfPageUp.Matches(event) {
		return tcell.NewEventKey(tcell.KeyPgUp, ' ', tcell.ModNone)
	}

	if keymap.LineStart.Matches(event) {
		return tcell.NewEventKey(tcell.KeyHome, ' ', tcell.ModNone)
	}

	if keymap.LineEnd.Matches(event) {
		return tcell.NewEventKey(tcell.KeyEnd, ' ', tcell.ModNone)
	}

	return event
}

func buildMainHelpText(keymap Keymap) string {
	menuKey := keymap.ToggleMenu.PrimaryString()
	if menuKey == "" {
		menuKey = "Ctrl-/"
	}

	submitKey := keymap.SubmitFilter.PrimaryString()
	if submitKey == "" {
		submitKey = "Enter"
	}

	return fmt.Sprintf("[::d]%s[::-] [::b]menu[::-]   [::d]Ctrl-C[::-] [::b]quit[::-]   [::d]%s[::-] [::b]quit and write output[::-]", menuKey, submitKey)
}

func appendKeybindingEntries(rows *[]overlay.KeybindingEntry, action string, bindings KeyBindings) {
	for _, binding := range bindings {
		*rows = append(*rows, overlay.KeybindingEntry{Action: action, Keybinding: binding.String()})
	}
}

func activeKeybindingEntries(keymap Keymap) []overlay.KeybindingEntry {
	rows := make([]overlay.KeybindingEntry, 0, 48)

	appendKeybindingEntries(&rows, "submit-filter", keymap.SubmitFilter)
	appendKeybindingEntries(&rows, "move-down", keymap.MoveDown)
	appendKeybindingEntries(&rows, "move-up", keymap.MoveUp)
	appendKeybindingEntries(&rows, "page-down", keymap.PageDown)
	appendKeybindingEntries(&rows, "line-start", keymap.LineStart)
	appendKeybindingEntries(&rows, "line-end", keymap.LineEnd)
	appendKeybindingEntries(&rows, "half-page-up", keymap.HalfPageUp)
	appendKeybindingEntries(&rows, "half-page-down", keymap.HalfPageDown)
	appendKeybindingEntries(&rows, "filter-cursor-right", keymap.FilterCursorRight)
	appendKeybindingEntries(&rows, "filter-cursor-left", keymap.FilterCursorLeft)
	appendKeybindingEntries(&rows, "focus-input-pane-up", keymap.FocusInputPaneUp)
	appendKeybindingEntries(&rows, "focus-input-pane-left", keymap.FocusInputPaneLeft)
	appendKeybindingEntries(&rows, "focus-output-pane", keymap.FocusOutputPane)
	appendKeybindingEntries(&rows, "focus-filter-input", keymap.FocusFilterInput)
	appendKeybindingEntries(&rows, "next-focus", keymap.NextFocus)
	appendKeybindingEntries(&rows, "previous-focus", keymap.PreviousFocus)
	appendKeybindingEntries(&rows, "toggle-input-pane", keymap.ToggleInputPane)
	appendKeybindingEntries(&rows, "save-filter-history", keymap.SaveFilterHistory)
	appendKeybindingEntries(&rows, "toggle-menu", keymap.ToggleMenu)
	appendKeybindingEntries(&rows, "textview-page-up", keymap.TextviewPageUp)
	appendKeybindingEntries(&rows, "textview-page-down", keymap.TextviewPageDown)
	appendKeybindingEntries(&rows, "textview-end", keymap.TextviewEnd)

	rows = append(rows,
		overlay.KeybindingEntry{Action: "quit", Keybinding: "Ctrl-C"},
	)

	return rows
}

func createApp(doc Document) *tview.Application {
	app := tview.NewApplication()

	// tview uses colors for a dark background by default, so reset some of
	// the styles to simply use the colors from the terminal to better
	// support light color themes
	tview.Styles.PrimaryTextColor = tcell.ColorDefault
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault
	tview.Styles.BorderColor = tcell.ColorDefault
	tview.Styles.TitleColor = tcell.ColorDefault
	tview.Styles.GraphicsColor = tcell.ColorDefault

	inputView := tview.NewTextView()
	inputView.SetDynamicColors(true).SetWrap(false).SetBorder(true)
	inputPane := pane{tv: inputView}

	outputView := tview.NewTextView()
	outputView.SetDynamicColors(true).SetWrap(false).SetBorder(true)
	outputPane := pane{tv: outputView}

	errorView := tview.NewTextView()
	errorView.SetDynamicColors(false).SetTitle("Error").SetBorder(true)

	helpView := tview.NewTextView()
	helpView.SetDynamicColors(true)
	helpView.SetTextAlign(tview.AlignCenter)
	helpView.SetText(buildMainHelpText(doc.config.Keymap))

	var filterHistory history
	filterHistory.Init(string(doc.options.HistoryFile))
	// If submit-filter includes Enter, we need SetDoneFunc to handle submission so
	// Enter still works with autocomplete selection.
	submitOnEnter := doc.config.Keymap.SubmitFilter.Matches(
		tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone),
	)
	submitFilter := func() {
		app.Stop()

		fmt.Fprintln(os.Stderr, doc.filter)

		// Enable or disable colors depending on if
		// stdout is a tty, respecting options set by
		// the user
		isTty := term.IsTerminal(int(os.Stdout.Fd()))
		if !isTty && !bool(doc.options.ForceColor) {
			doc.options.Monochrome = true
		} else if isTty && !bool(doc.options.Monochrome) {
			doc.options.ForceColor = true
		}

		filterHistory.Add(doc.filter)

		if _, err := doc.WriteTo(os.Stdout); err != nil {
			log.Fatalln(err)
		}
	}

	var (
		mutex  sync.Mutex
		cancel context.CancelFunc = func() {}
	)

	cond := sync.NewCond(&mutex)

	// Initialize pending to true so that the output pane will update with the initial filter
	pending := true

	// Create a cancellable context when writing to the output view. If the
	// filter input changes, the context is cancelled and the process is
	// killed. This must be set before filterInput is created because
	// tview's SetAutocompleteFunc triggers an initial autocomplete that
	// spawns a goroutine reading doc.
	doc.ctx, cancel = context.WithCancel(context.Background())

	filterMap := make(map[string][]string)
	queueDocumentUpdate := func(update func(*Document)) {
		mutex.Lock()
		defer mutex.Unlock()

		cancel()
		update(&doc)
		pending = true
		cond.Signal()
	}

	filterInput := tview.NewInputField()
	filterInput.
		SetText(doc.filter).
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldTextColor(tcell.ColorDefault).
		SetChangedFunc(func(text string) {
			errorView.Clear()
			filterInput.SetFieldTextColor(tcell.ColorDefault)

			if text == doc.filter {
				return
			}

			queueDocumentUpdate(func(next *Document) {
				*next = next.WithFilter(text)
			})
		}).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter && submitOnEnter {
				submitFilter()
			}
		}).
		SetAutocompleteFunc(func(text string) []string {
			if text == "" {
				return filterHistory.Entries()
			}

			if pos := strings.LastIndexByte(text, '.'); pos != -1 {
				prefix := text[0:pos]
				trimmed := strings.TrimSpace(prefix)

				mutex.Lock()
				candidates, ok := filterMap[trimmed]
				mutex.Unlock()
				if ok {
					cur := text[pos+1:]
					var entries []string
					for _, c := range candidates {
						key := c[pos+1:]
						if strings.HasPrefix(key, cur) {
							entries = append(entries, c)
						}
					}

					return entries
				}

				go func() {
					var filt string
					if prefix != "" {
						p, _ := strings.CutSuffix(trimmed, "|")
						filt = p + "| keys"
					} else {
						filt = "keys"
					}

					mutex.Lock()
					filtered := doc.WithFilter("[" + filt + "] | unique | first")
					mutex.Unlock()

					var buf bytes.Buffer
					_, err := filtered.WriteTo(&buf)
					if err != nil {
						return
					}

					var keys []string
					if err := json.Unmarshal(buf.Bytes(), &keys); err != nil {
						return
					}

					entries := keys[:0]
					for _, k := range keys {
						if k == "" {
							k = `""`
						} else {
							first := strings.ToLower(string(k[0]))
							if strings.ContainsAny(k, specialChars) || !strings.Contains(alphabet, first) {
								k = `"` + k + `"`
							}
						}

						entries = append(entries, prefix+"."+k)
					}

					mutex.Lock()
					filterMap[trimmed] = entries
					mutex.Unlock()

					filterInput.Autocomplete()

					app.Draw()
				}()
			}

			return nil
		}).
		SetAutocompleteUseTags(false).
		SetAutocompleteStyles(tcell.ColorBlack, tcell.StyleDefault.Background(tcell.ColorBlack), tcell.StyleDefault.Reverse(true)).
		SetTitle("Filter").
		SetBorder(true)

	saveCurrentFilterToHistory := func() (status string, expression string, err error) {
		expression = strings.TrimSpace(filterInput.GetText())
		if expression == "" {
			return "empty", expression, nil
		}

		if filterHistory.path == "" {
			return "disabled", expression, nil
		}

		added, err := filterHistory.AddIfMissing(expression)
		if err != nil {
			return "", expression, err
		}

		if added {
			return "added", expression, nil
		}

		return "exists", expression, nil
	}

	// Initialize the initial line counts to some large number. If the
	// input is small, this will be updated to the correct value before it
	// is ever displayed in the UI. But for large inputs (which will take
	// longer to calculate the correct value), this is a better initial
	// guess.
	var inputLineCount atomic.Int64
	inputLineCount.Store(10000)
	var outputLineCount atomic.Int64
	outputLineCount.Store(10000)

	// Process document with empty filter to populate input view
	go func() {
		_, err := doc.WithFilter(".").WriteTo(&inputPane)
		if err != nil {
			log.Printf("Error while running jq on input: %s\n", err)
			return
		}

		inputLineCount.Store(int64(strings.Count(inputView.GetText(false), "\n")))
	}()

	go func() {
		for {
			cond.L.Lock()
			for !pending {
				cond.Wait()
			}

			d := doc
			pending = false

			// Re-initialize the cancellable context while still holding the
			// lock so that concurrent calls to cancel() in
			// queueDocumentUpdate always operate on a fully constructed
			// context.
			d.ctx, cancel = context.WithCancel(context.Background())
			cond.L.Unlock()

			_, err := d.WriteTo(&outputPane)
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					if code := exitErr.ExitCode(); code != -1 {
						app.QueueUpdate(func() {
							filterInput.SetFieldTextColor(tcell.ColorMaroon)
							fmt.Fprint(tview.ANSIWriter(errorView), string(exitErr.Stderr))
						})
					}
				}
			} else {
				outputLineCount.Store(int64(strings.Count(outputView.GetText(false), "\n")))
			}

			app.Draw()
		}
	}()

	inputPaneProportion := 1
	if doc.options.HideInputPane {
		inputPaneProportion = 0
	}
	viewFlex := tview.NewFlex().
		AddItem(inputView, 0, inputPaneProportion, false).
		AddItem(outputView, 0, 1, false)
	grid := tview.NewGrid().
		SetRows(0, 3, 4, 1).
		SetColumns(0).
		AddItem(viewFlex, 0, 0, 1, 1, 0, 0, false).
		AddItem(tview.NewFlex().
			AddItem(tview.NewBox(), 0, 1, false).
			AddItem(filterInput, 0, 4, true).
			AddItem(tview.NewBox(), 0, 1, false), 1, 0, 1, 1, 0, 0, true).
		AddItem(tview.NewFlex().
			AddItem(tview.NewBox(), 0, 1, false).
			AddItem(errorView, 0, 4, false).
			AddItem(tview.NewBox(), 0, 1, false), 2, 0, 1, 1, 0, 0, false).
		AddItem(helpView, 3, 0, 1, 1, 0, 0, false)

	pages := tview.NewPages().
		AddPage("main", grid, true, true)

	historyNotice := tview.NewTextView()
	historyNotice.SetBorder(true)
	historyNotice.SetTitle("History")
	historyNotice.SetTextAlign(tview.AlignCenter)

	historyNoticeContainer := tview.NewGrid().
		SetRows(0, 3, 0).
		SetColumns(0, 24, 0).
		AddItem(historyNotice, 1, 1, 1, 1, 0, 0, true)

	const historyNoticePage = "history-notice"
	pages.AddPage(historyNoticePage, historyNoticeContainer, true, false)

	isHistoryNoticeOpen := false
	showHistoryNotice := func(message string) {
		historyNotice.SetText(message)

		width := max(tview.TaggedStringWidth(message)+4, 24)

		historyNoticeContainer.SetColumns(0, width, 0)
		historyNoticeContainer.SetRows(0, 3, 0)
		pages.ShowPage(historyNoticePage)
		pages.SendToFront(historyNoticePage)
		isHistoryNoticeOpen = true
	}

	closeHistoryNotice := func() {
		if !isHistoryNoticeOpen {
			return
		}

		pages.HidePage(historyNoticePage)
		isHistoryNoticeOpen = false
	}

	overlayPopup := overlay.NewController(app, pages, "overlay", overlay.Callbacks{
		ConfigureRows: func() []string { return overlay.ConfigureRows(doc.options) },
		ToggleConfigureRow: func(option options.Option) {
			switch option.(type) {
			case *options.HideInputPane:
				// This option only affects the ijq UI, not jq
				// itself, so we handle it differently
				mutex.Lock()
				doc.options.HideInputPane = !doc.options.HideInputPane
				hidden := doc.options.HideInputPane
				mutex.Unlock()
				if hidden {
					if inputView.HasFocus() {
						app.SetFocus(outputView)
					}
					viewFlex.ResizeItem(inputView, 0, 0)
				} else {
					viewFlex.ResizeItem(inputView, 0, 1)
				}
			default:
				queueDocumentUpdate(func(next *Document) {
					next.options.Toggle(option)
				})
			}
		},
		SaveCurrentFilterToHistory: func() (string, error) {
			status, _, err := saveCurrentFilterToHistory()
			if err != nil {
				return "", err
			}

			switch status {
			case "added":
				return "saved", nil
			case "exists":
				return "already in history", nil
			case "empty":
				return "filter is empty", nil
			case "disabled":
				return "history disabled", nil
			default:
				return "", nil
			}
		},
		LoadHistoryEntries: func() []string {
			return filterHistory.Entries()
		},
		DeleteHistoryEntryAt: func(index int) error {
			return filterHistory.DeleteAt(index)
		},
		ApplyHistoryEntry: func(expression string) {
			errorView.Clear()
			filterInput.SetFieldTextColor(tcell.ColorDefault)
			filterInput.SetText(expression)
		},
		ActiveKeybindings: func() []overlay.KeybindingEntry {
			return activeKeybindingEntries(doc.config.Keymap)
		},
	})

	pages.AddPage("overlay", overlayPopup.Primitive(), true, false)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		focused := app.GetFocus()
		activeKeymaps := doc.config.Keymap

		if isHistoryNoticeOpen {
			closeHistoryNotice()
			return nil
		}

		if overlayPopup.IsOpen() {
			if activeKeymaps.ToggleMenu.Matches(event) {
				overlayPopup.Close()
				return nil
			}

			event = normalizeOverlayEvent(event, activeKeymaps)
			return overlayPopup.HandleInput(event)
		}

		if filterInput.HasFocus() {
			if event.Key() == tcell.KeyEnter && event.Modifiers() == tcell.ModNone {
				// Let tview process Enter first so autocomplete selections work.
				return event
			}

			if activeKeymaps.SubmitFilter.Matches(event) {
				submitFilter()
				return nil
			}

			if event.Key() == tcell.KeyRune && event.Modifiers() == tcell.ModNone {
				// Keep printable characters available for typing in the filter field,
				// even if they are used as key bindings in other contexts.
				return event
			}
		}

		if activeKeymaps.ToggleMenu.Matches(event) {
			overlayPopup.Open()
			return nil
		}

		if activeKeymaps.SaveFilterHistory.Matches(event) {
			status, expression, err := saveCurrentFilterToHistory()
			if err != nil {
				showHistoryNotice("Failed to save filter to history")
				return nil
			}

			switch status {
			case "added":
				expression = strings.ReplaceAll(expression, "\n", " ")
				showHistoryNotice(fmt.Sprintf("Added %s to history", expression))
			case "exists":
				showHistoryNotice("Filter already in history")
			case "empty":
				showHistoryNotice("Filter is empty")
			case "disabled":
				showHistoryNotice("History is disabled")
			}

			return nil
		}

		if event.Key() == tcell.KeyCtrlC {
			app.Stop()
			return nil
		}

		if activeKeymaps.MoveDown.Matches(event) {
			return tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone)
		}

		if activeKeymaps.MoveUp.Matches(event) {
			return tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModNone)
		}

		if activeKeymaps.PageDown.Matches(event) {
			return tcell.NewEventKey(tcell.KeyPgDn, ' ', tcell.ModNone)
		}

		if activeKeymaps.LineStart.Matches(event) {
			if tv, ok := focused.(*tview.TextView); ok {
				scrollHorizontally(tv, false)
				return nil
			}
		}

		if activeKeymaps.LineEnd.Matches(event) {
			if tv, ok := focused.(*tview.TextView); ok {
				scrollHorizontally(tv, true)
				return nil
			}
		}

		if activeKeymaps.HalfPageUp.Matches(event) {
			if tv, ok := focused.(*tview.TextView); ok {
				scrollHalfPage(tv, true)
				return nil
			}
		}

		if activeKeymaps.HalfPageDown.Matches(event) {
			if tv, ok := focused.(*tview.TextView); ok {
				scrollHalfPage(tv, false)
				return nil
			}
		}

		if activeKeymaps.FilterCursorRight.Matches(event) {
			if filterInput.HasFocus() {
				return tcell.NewEventKey(tcell.KeyRight, ' ', tcell.ModNone)
			}
		}

		if activeKeymaps.FilterCursorLeft.Matches(event) {
			if filterInput.HasFocus() {
				return tcell.NewEventKey(tcell.KeyLeft, ' ', tcell.ModNone)
			}
		}

		if activeKeymaps.FocusInputPaneUp.Matches(event) {
			if filterInput.HasFocus() {
				if !doc.options.HideInputPane {
					app.SetFocus(inputView)
				} else {
					app.SetFocus(outputView)
				}
				return nil
			}
		}

		if activeKeymaps.FocusInputPaneLeft.Matches(event) {
			if !doc.options.HideInputPane {
				app.SetFocus(inputView)
				return nil
			}
		}

		if activeKeymaps.FocusOutputPane.Matches(event) {
			app.SetFocus(outputView)
			return nil
		}

		if activeKeymaps.FocusFilterInput.Matches(event) {
			app.SetFocus(filterInput)
			return nil
		}

		if activeKeymaps.NextFocus.Matches(event) {
			if inputView.HasFocus() {
				app.SetFocus(outputView)
				return nil
			}

			if outputView.HasFocus() {
				app.SetFocus(filterInput)
				return nil
			}

			if filterInput.HasFocus() {
				return tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone)
			}
		}

		if activeKeymaps.PreviousFocus.Matches(event) {
			if inputView.HasFocus() {
				app.SetFocus(filterInput)
				return nil
			}

			if outputView.HasFocus() {
				app.SetFocus(inputView)
				return nil
			}

			if filterInput.HasFocus() {
				return tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModNone)
			}
		}

		if activeKeymaps.ToggleInputPane.Matches(event) {
			mutex.Lock()
			hidden := doc.options.HideInputPane
			mutex.Unlock()

			if !hidden {
				if inputView.HasFocus() {
					app.SetFocus(outputView)
				}

				viewFlex.ResizeItem(inputView, 0, 0)
				mutex.Lock()
				doc.options.HideInputPane = true
				mutex.Unlock()
				return nil
			}

			viewFlex.ResizeItem(inputView, 0, 1)
			mutex.Lock()
			doc.options.HideInputPane = false
			mutex.Unlock()
			return nil
		}

		if tv, ok := focused.(*tview.TextView); ok {
			if activeKeymaps.TextviewPageUp.Matches(event) {
				return tcell.NewEventKey(tcell.KeyCtrlB, ' ', tcell.ModNone)
			}

			if activeKeymaps.TextviewPageDown.Matches(event) {
				return tcell.NewEventKey(tcell.KeyCtrlF, ' ', tcell.ModNone)
			}

			if activeKeymaps.TextviewEnd.Matches(event) {
				// tview handles G natively but does not
				// redraw, so the scroll indicator doesn't
				// update. So we handle G ourselves and force a
				// redraw
				tv.ScrollToEnd()
				app.ForceDraw()
				return nil
			}
		}

		return event
	})

	app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		// Start a synchronized update
		tty, ok := screen.Tty()
		if ok {
			tty.Write([]byte("\x1b[?2026h"))
		}

		updateScrollIndicator("Input", int(inputLineCount.Load()), inputView)
		updateScrollIndicator("Output", int(outputLineCount.Load()), outputView)

		return false
	})

	app.SetAfterDrawFunc(func(screen tcell.Screen) {
		// Finish a synchronized update
		tty, ok := screen.Tty()
		if ok {
			tty.Write([]byte("\x1b[?2026l"))
		}
	})

	app.SetRoot(pages, true).EnableMouse(true).SetFocus(grid)

	return app
}

func main() {
	// Remove log prefix
	log.SetFlags(0)

	configPath, err := DefaultConfigPath()
	if err != nil {
		log.Fatalf("error getting default config path: %s\n", err)
	}

	config, err := NewConfig(configPath)
	if err != nil {
		log.Fatalf("error loading config file %q: %s\n", configPath, err)
	}

	options := options.Options{
		HistoryFile:   config.HistoryFile,
		JQCommand:     config.JQCommand,
		HideInputPane: config.HideInputPane,
		LibraryPaths:  config.LibraryPaths,
	}

	filter, args := parseArgs(&options)

	if _, err := exec.LookPath(string(options.JQCommand)); err != nil {
		log.Fatalf("%s is not installed or could not be found: %s\n", options.JQCommand, err)
	}

	doc := Document{filter: filter, options: options, config: config}

	if !options.NullInput {
		var in io.Reader = os.Stdin
		if len(args) > 0 {
			var files []io.Reader
			for _, fname := range args {
				f, err := os.Open(fname)
				if err != nil {
					log.Fatalln(err)
				}

				defer f.Close()

				files = append(files, f)
			}

			in = io.MultiReader(files...)
		}

		if _, err := doc.ReadFrom(in); err != nil {
			log.Fatalln(err)
		}
	}

	app := createApp(doc)
	if err := app.Run(); err != nil {
		log.Fatalln(err)
	}
}
