// Copyright (C) 2020 Gregory Anders
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
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var Version string

type Options struct {
	compact    bool
	nullInput  bool
	slurp      bool
	rawOutput  bool
	rawInput   bool
	monochrome bool
	sortKeys   bool
}

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

	if !o.monochrome {
		opts = append(opts, "-C")
	}

	if o.sortKeys {
		opts = append(opts, "-S")
	}

	return opts
}

func stdinHasData() bool {
	stat, _ := os.Stdin.Stat()
	return stat.Mode()&os.ModeCharDevice == 0
}

type Document struct {
	contents string
	options  Options
}

func (d *Document) FromFile(filename string) error {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	d.contents = string(bytes)
	return nil
}

func (d *Document) FromStdin() error {
	if !stdinHasData() {
		// stdin is not being piped
		return errors.New("No data on stdin")
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(os.Stdin); err != nil {
		return err
	}

	d.contents = buf.String()

	return nil
}

func (d *Document) Read(args []string) error {
	if d.options.nullInput {
		return nil
	}

	if len(args) > 0 {
		for _, file := range args {
			if err := d.FromFile(file); err != nil {
				return err
			}
		}
	} else {
		if err := d.FromStdin(); err != nil {
			return err
		}
	}

	return nil
}

func (d *Document) Filter(filter string) (string, error) {
	args := append(d.options.ToSlice(), filter)
	cmd := exec.Command("jq", args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	go func() {
		defer stdin.Close()
		_, _ = io.WriteString(stdin, d.contents);
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(out), nil

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
	flag.BoolVar(&options.monochrome, "M", false, "don't colorize JSON")
	flag.BoolVar(&options.sortKeys, "S", false, "sort keys of objects on output")

	filterFile := flag.String("f", "", "read initial filter from `filename`")
	version := flag.Bool("V", false, "print version and exit")

	flag.Parse()

	if *version {
		fmt.Println("ijq " + Version)
		os.Exit(0)
	}

	filter := "."
	args := flag.Args()

	if *filterFile != "" {
		contents, err := ioutil.ReadFile(*filterFile)
		if err != nil {
			log.Fatalln(err)
		}

		filter = string(contents)
	} else if len(args) > 1 || (len(args) > 0 && (stdinHasData() || options.nullInput)) {
		filter = args[0]
		args = args[1:]
	} else if len(args) == 0 && !stdinHasData() && !options.nullInput {
		flag.Usage()
		os.Exit(1)
	}

	return options, filter, args
}

func createApp(doc Document, filter string) *tview.Application {
	app := tview.NewApplication()

	inputView := tview.NewTextView().SetDynamicColors(true)
	inputView.SetTitle("Input").SetBorder(true)

	outputView := tview.NewTextView().
		SetDynamicColors(true).
		SetChangedFunc(func() {
			app.Draw()
		})

	outputView.SetTitle("Output").SetBorder(true)

	outputWriter := tview.ANSIWriter(outputView)

	filterInput := tview.NewInputField()
	filterInput.
		SetText(filter).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldTextColor(tcell.ColorSilver).
		SetChangedFunc(func(text string) {
				go app.QueueUpdate(func() {
					out, err := doc.Filter(text)
					if err != nil {
						filterInput.SetFieldTextColor(tcell.ColorMaroon)
						return
					}

					filterInput.SetFieldTextColor(tcell.ColorSilver)
					outputView.Clear()
					fmt.Fprint(outputWriter, out)
					outputView.ScrollToBeginning()
				})
		}).
		SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyEnter:
				app.Stop()
				fmt.Fprintln(os.Stderr, filterInput.GetText())
				fmt.Fprint(os.Stdout, outputView.GetText(true))
			}
		})

	filterInput.SetTitle("Filter").SetBorder(true)

	// Filter output with original filter
	go func() {
		orig, err := doc.Filter(".")
		if err != nil {
			log.Fatalln(err)
		}

		out, err := doc.Filter(filter)
		if err != nil {
			filterInput.SetFieldTextColor(tcell.ColorMaroon)
		}

		fmt.Fprint(tview.ANSIWriter(inputView), orig)
		fmt.Fprint(outputWriter, out)
	}()

	grid := tview.NewGrid().
		SetRows(0, 3).
		SetColumns(0).
		AddItem(tview.NewFlex().
			AddItem(inputView, 0, 1, false).
			AddItem(outputView, 0, 1, false), 0, 0, 1, 1, 0, 0, false).
		AddItem(tview.NewFlex().
			AddItem(tview.NewBox(), 0, 1, false).
			AddItem(filterInput, 0, 3, true).
			AddItem(tview.NewBox(), 0, 1, false), 1, 0, 1, 1, 0, 0, true)

	elements := []tview.Primitive{inputView, outputView, filterInput}
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		var off int
		switch key := event.Key(); key {
		case tcell.KeyTab:
			off = 1
		case tcell.KeyBacktab:
			off = -1
		default:
			return event
		}

		for i, e := range elements {
			if e.HasFocus() {
				if i+off < len(elements) && i+off >= 0 {
					app.SetFocus(elements[i+off])
				} else if i+off == len(elements) {
					app.SetFocus(elements[0])
				} else {
					app.SetFocus(elements[len(elements)-1])
				}
				return nil
			}
		}

		return event
	})

	app.SetRoot(grid, true).SetFocus(grid)

	return app
}

func main() {
	// Remove log prefix
	log.SetFlags(0)

	options, filter, args := parseArgs()

	doc := Document{options: options}
	if err := doc.Read(args); err != nil {
		log.Fatalln(err)
	}

	app := createApp(doc, filter)
	if err := app.Run(); err != nil {
		log.Fatalln(err)
	}
}
