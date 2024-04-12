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
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/kyoh86/xdg"
	"github.com/rivo/tview"
	"golang.org/x/term"
)

const DefaultCommand string = "jq"

// Special characters that, if present in a JSON key, need to be quoted in the
// jq filter
const SpecialChars string = ".-:$/"

const Alphabet string = "abcdefghijklmnopqrstuvwxyz"

var Version string

type Options struct {
	compact     bool
	command     string
	nullInput   bool
	slurp       bool
	rawOutput   bool
	rawInput    bool
	monochrome  bool
	sortKeys    bool
	historyFile string
	forceColor  bool
}

// Convert the Options struct to a string slice of option flags that gets
// passed to jq.
func (o *Options) ToSlice() []string {
	opts := []string{}

	if o.compact {
		opts = append(opts, "-c")
	}

	if o.nullInput {
		opts = append(opts, "-n")
	}

	if o.slurp {
		opts = append(opts, "-s")
	}

	if o.rawOutput {
		opts = append(opts, "-r")
	}

	if o.rawInput {
		opts = append(opts, "-R")
	}

	if o.monochrome {
		opts = append(opts, "-M")
	}

	if o.forceColor {
		opts = append(opts, "-C")
	}

	if o.sortKeys {
		opts = append(opts, "-S")
	}

	return opts
}

type Document struct {
	input   string
	filter  string
	options Options
}

func (d Document) WithFilter(filter string) Document {
	return Document{input: d.input, filter: filter, options: d.options}
}

func (d *Document) ReadFrom(r io.Reader) (n int64, err error) {
	var buf bytes.Buffer
	n, err = buf.ReadFrom(r)
	d.input = buf.String()
	return n, err
}

func (d Document) WriteTo(w io.Writer) (n int64, err error) {
	opts := d.options
	if _, ok := w.(*tview.TextView); ok {
		// Writer is a TextView, so set options accordingly
		opts.forceColor = true
		opts.monochrome = false
		opts.compact = false
		opts.rawOutput = false
	}

	args := append(opts.ToSlice(), d.filter)
	cmd := exec.Command(d.options.command, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return 0, err
	}

	go func() {
		defer stdin.Close()
		_, _ = io.WriteString(stdin, d.input)
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// jq prints its error message to standard out, but we
			// will deliver it in the Stderr field as this will
			// most likely be an exec.ExitError.
			exiterr.Stderr = out
		}
		return 0, err
	}

	if tv, ok := w.(*tview.TextView); ok {
		w = tview.ANSIWriter(tv)
		tv.Clear()
	}

	m, err := w.Write(out)
	n = int64(m)
	return n, err
}

func parseArgs() (Options, string, []string) {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "ijq - interactive jq\n\n")
		fmt.Fprintf(os.Stderr, "Usage: ijq [-cnsrRMSV] [-f file] [filter] [files ...]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	options := Options{}
	flag.BoolVar(&options.compact, "c", false, "compact instead of pretty-printed output")
	flag.BoolVar(&options.nullInput, "n", false, "use ```null` as the single input value")
	flag.BoolVar(&options.slurp, "s", false, "read (slurp) all inputs into an array; apply filter to it")
	flag.BoolVar(&options.rawOutput, "r", false, "output raw strings, not JSON texts")
	flag.BoolVar(&options.rawInput, "R", false, "read raw strings, not JSON texts")
	flag.BoolVar(&options.forceColor, "C", false, "force colorized JSON, even if writing to a pipe or file")
	flag.BoolVar(&options.monochrome, "M", false, "monochrome (don't colorize JSON)")
	flag.BoolVar(&options.sortKeys, "S", false, "sort keys of objects on output")

	flag.StringVar(
		&options.command,
		"jqbin",
		DefaultCommand,
		"name of or path to jq binary to use",
	)

	flag.StringVar(
		&options.historyFile,
		"H",
		filepath.Join(xdg.DataHome(), "ijq", "history"),
		"set path to history file. Set to '' to disable history.",
	)

	filterFile := flag.String("f", "", "read initial filter from `filename`")
	version := flag.Bool("V", false, "print version and exit")

	flag.Parse()

	if *version {
		fmt.Println("ijq " + Version)
		os.Exit(0)
	}

	filter := "."
	args := flag.Args()

	stdinIsTty := term.IsTerminal(int(os.Stdin.Fd()))

	if *filterFile != "" {
		contents, err := os.ReadFile(*filterFile)
		if err != nil {
			log.Fatalln(err)
		}

		filter = string(contents)
	} else if len(args) > 1 || (len(args) > 0 && (!stdinIsTty || options.nullInput)) {
		filter = args[0]
		args = args[1:]
	} else if len(args) == 0 && stdinIsTty && !options.nullInput {
		flag.Usage()
		os.Exit(1)
	}

	return options, filter, args
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

	outputView := tview.NewTextView()
	outputView.SetDynamicColors(true).SetWrap(false).SetBorder(true)
	outputView.SetChangedFunc(func() {
		outputView.ScrollToBeginning()
		app.Draw()
	})

	errorView := tview.NewTextView()
	errorView.SetDynamicColors(true).SetTitle("Error").SetBorder(true)

	var filterHistory history
	filterHistory.Init(doc.options.historyFile)

	var inputLineCount int
	var outputLineCount int

	var mutex sync.Mutex
	cond := sync.NewCond(&mutex)
	pending := false

	filterMap := make(map[string][]string)
	filterInput := tview.NewInputField()
	filterInput.
		SetText(doc.filter).
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldTextColor(tcell.ColorDefault).
		SetChangedFunc(func(text string) {
			errorView.Clear()
			mutex.Lock()
			defer mutex.Unlock()
			doc = doc.WithFilter(text)
			pending = true
			cond.Signal()
		}).
		SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyEnter:
				app.Stop()

				fmt.Fprintln(os.Stderr, doc.filter)

				// Enable or disable colors depending on if
				// stdout is a tty, respecting options set by
				// the user
				isTty := term.IsTerminal(int(os.Stdout.Fd()))
				if !isTty && !doc.options.forceColor {
					doc.options.monochrome = true
				} else if isTty && !doc.options.monochrome {
					doc.options.forceColor = true
				}

				filterHistory.Add(doc.filter)

				if _, err := doc.WriteTo(os.Stdout); err != nil {
					log.Fatalln(err)
				}
			}
		}).
		SetAutocompleteFunc(func(text string) []string {
			if text == "" {
				var entries []string
				for _, item := range filterHistory.Items {
					entries = append(entries, tview.Escape(item))
				}
				return entries
			}

			if pos := strings.LastIndexByte(text, '.'); pos != -1 {
				prefix := text[0:pos]

				mutex.Lock()
				candidates, ok := filterMap[prefix]
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
						filt = prefix + "| keys"
					} else {
						filt = "keys"
					}

					var buf bytes.Buffer
					_, err := doc.WithFilter("[" + filt + "] | unique | first").WriteTo(&buf)
					if err != nil {
						return
					}

					var keys []string
					if err := json.Unmarshal(buf.Bytes(), &keys); err != nil {
						return
					}

					entries := keys[:0]
					for _, k := range keys {
						first := strings.ToLower(string(k[0]))
						if strings.ContainsAny(k, SpecialChars) || !strings.Contains(Alphabet, first) {
							k = `"` + k + `"`
						}
						entries = append(entries, prefix+"."+k)
					}

					mutex.Lock()
					filterMap[prefix] = entries
					mutex.Unlock()

					filterInput.Autocomplete()

					app.Draw()
				}()
			}

			return nil
		}).
		SetAutocompleteStyles(tcell.ColorBlack, tcell.StyleDefault, tcell.StyleDefault.Reverse(true)).
		SetTitle("Filter").
		SetBorder(true)

	// Generate formatted input and output with original filter
	go app.QueueUpdateDraw(func() {
		if _, err := doc.WithFilter(".").WriteTo(inputView); err != nil {
			log.Fatalln(err)
		}

		outputView.ScrollToBeginning()
		if _, err := doc.WriteTo(outputView); err != nil {
			filterInput.SetFieldTextColor(tcell.ColorMaroon)
		}

		inputLineCount = strings.Count(inputView.GetText(false), "\n")
		outputLineCount = strings.Count(outputView.GetText(false), "\n")
	})

	go func() {
		for {
			mutex.Lock()
			for !pending {
				cond.Wait()
			}

			d := doc
			pending = false
			mutex.Unlock()

			_, err := d.WriteTo(outputView)
			if err != nil {
				filterInput.SetFieldTextColor(tcell.ColorMaroon)
				exitErr, ok := err.(*exec.ExitError)
				if ok {
					fmt.Fprint(tview.ANSIWriter(errorView), string(exitErr.Stderr))
				}
			} else {
				outputLineCount = strings.Count(outputView.GetText(false), "\n")
				filterInput.SetFieldTextColor(tcell.ColorDefault)
			}
		}
	}()

	grid := tview.NewGrid().
		SetRows(0, 3, 4).
		SetColumns(0).
		AddItem(tview.NewFlex().
			AddItem(inputView, 0, 1, false).
			AddItem(outputView, 0, 1, false), 0, 0, 1, 1, 0, 0, false).
		AddItem(tview.NewFlex().
			AddItem(tview.NewBox(), 0, 1, false).
			AddItem(filterInput, 0, 4, true).
			AddItem(tview.NewBox(), 0, 1, false), 1, 0, 1, 1, 0, 0, true).
		AddItem(tview.NewFlex().
			AddItem(tview.NewBox(), 0, 1, false).
			AddItem(errorView, 0, 4, false).
			AddItem(tview.NewBox(), 0, 1, false), 2, 0, 1, 1, 0, 0, false)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		shift := event.Modifiers()&tcell.ModShift != 0
		focused := app.GetFocus()

		switch key := event.Key(); key {
		case tcell.KeyCtrlN:
			return tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone)
		case tcell.KeyCtrlP:
			return tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModNone)
		case tcell.KeyCtrlV:
			return tcell.NewEventKey(tcell.KeyPgDn, ' ', tcell.ModNone)
		case tcell.KeyCtrlA:
			if tv, ok := focused.(*tview.TextView); ok {
				scrollHorizontally(tv, false)
				return nil
			}
		case tcell.KeyCtrlE:
			if tv, ok := focused.(*tview.TextView); ok {
				scrollHorizontally(tv, true)
				return nil
			}
		case tcell.KeyCtrlU:
			if tv, ok := focused.(*tview.TextView); ok {
				scrollHalfPage(tv, true)
				return nil
			}
		case tcell.KeyCtrlD:
			if tv, ok := focused.(*tview.TextView); ok {
				scrollHalfPage(tv, false)
				return nil
			}
		case tcell.KeyUp:
			if shift && filterInput.HasFocus() {
				app.SetFocus(inputView)
				return nil
			}
		case tcell.KeyLeft:
			if shift {
				app.SetFocus(inputView)
				return nil
			}
		case tcell.KeyRight:
			if shift {
				app.SetFocus(outputView)
				return nil
			}
		case tcell.KeyDown:
			if shift {
				app.SetFocus(filterInput)
				return nil
			}
		case tcell.KeyTab:
			if inputView.HasFocus() {
				app.SetFocus(outputView)
				return nil
			} else if outputView.HasFocus() {
				app.SetFocus(filterInput)
				return nil
			} else if filterInput.HasFocus() {
				return tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone)
			}
		case tcell.KeyBacktab:
			if inputView.HasFocus() {
				app.SetFocus(filterInput)
				return nil
			} else if outputView.HasFocus() {
				app.SetFocus(inputView)
				return nil
			} else if filterInput.HasFocus() {
				return tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModNone)
			}
		}

		if tv, ok := focused.(*tview.TextView); ok {
			switch ru := event.Rune(); ru {
			case '0':
				scrollHorizontally(tv, false)
				return nil
			case '$':
				scrollHorizontally(tv, true)
				return nil
			case 'd':
				scrollHalfPage(tv, false)
				return nil
			case 'u':
				scrollHalfPage(tv, true)
				return nil
			case 'b':
				return tcell.NewEventKey(tcell.KeyCtrlB, ' ', tcell.ModNone)
			case 'f':
				return tcell.NewEventKey(tcell.KeyCtrlF, ' ', tcell.ModNone)
			case 'v':
				if event.Modifiers()&tcell.ModAlt != 0 {
					return tcell.NewEventKey(tcell.KeyPgUp, ' ', tcell.ModNone)
				}
			case 'G':
				// tview handles G natively but does not
				// redraw, so the scroll indicator doesn't
				// update. So we handle G ourselves and force a
				// redraw
				tv.ScrollToEnd()
				app.ForceDraw()
			}
		}

		return event
	})

	app.SetBeforeDrawFunc(func(_ tcell.Screen) bool {
		updateScrollIndicator("Input", inputLineCount, inputView)
		updateScrollIndicator("Output", outputLineCount, outputView)
		return false
	})

	app.SetRoot(grid, true).EnableMouse(true).SetFocus(grid)

	return app
}

func main() {
	// Remove log prefix
	log.SetFlags(0)

	options, filter, args := parseArgs()

	if _, err := exec.LookPath(options.command); err != nil {
		log.Fatalf("%s is not installed or could not be found: %s\n", options.command, err)
	}

	doc := Document{filter: filter, options: options}

	if !options.nullInput {
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
