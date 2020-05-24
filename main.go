package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/gdamore/tcell"
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
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		// stdin is not being piped
		return errors.New("No data on stdin")
	}

	scanner := bufio.NewScanner(os.Stdin)
	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	d.contents = strings.Join(lines, "\n")

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
		io.WriteString(stdin, d.contents)
	}()

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil

}

func main() {
	options := Options{}
	flag.BoolVar(&options.compact, "c", false, "compact instead of pretty-printed output")
	flag.BoolVar(&options.nullInput, "n", false, "use ```null` as the single input value")
	flag.BoolVar(&options.slurp, "s", false, "read (slurp) all inputs into an array; apply filter to it")
	flag.BoolVar(&options.rawOutput, "r", false, "output raw strings, not JSON texts")
	flag.BoolVar(&options.rawInput, "R", false, "read raw strings, not JSON texts")
	flag.BoolVar(&options.monochrome, "M", false, "don't colorize JSON")
	flag.BoolVar(&options.sortKeys, "S", false, "sort keys of objects on output")

	filter := flag.String("f", ".", "initial filter")
	version := flag.Bool("V", false, "print version and exit")

	flag.Parse()

	if *version {
		fmt.Println("ijq " + Version)
		return
	}

	app := tview.NewApplication()

	originalView := tview.NewTextView().SetDynamicColors(true)
	originalView.SetTitle("Original").SetBorder(true)

	outputView := tview.NewTextView().
		SetDynamicColors(true).
		SetChangedFunc(func() {
			app.Draw()
		})

	outputView.SetTitle("Output").SetBorder(true)

	outputWriter := tview.ANSIWriter(outputView)

	doc := Document{options: options}
	go func() {
		if flag.Arg(0) != "" {
			if err := doc.FromFile(flag.Arg(0)); err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
		} else {
			if err := doc.FromStdin(); err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
		}

		out, err := doc.Filter(*filter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}

		fmt.Fprint(tview.ANSIWriter(originalView), out)
		fmt.Fprint(outputWriter, out)
	}()

	filterInput := tview.NewInputField()
	filterInput.
		SetText(*filter).
		SetFieldBackgroundColor(0).
		SetFieldTextColor(7).
		SetChangedFunc(func(text string) {
			go func() {
				out, err := doc.Filter(text)
				if err != nil {
					filterInput.SetFieldTextColor(1)
					return
				}

				filterInput.SetFieldTextColor(7)
				outputView.Clear()
				fmt.Fprint(outputWriter, out)
				outputView.ScrollToBeginning()
			}()
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

	grid := tview.NewGrid().
		SetRows(0, 3).
		SetColumns(0).
		AddItem(tview.NewFlex().
			AddItem(outputView, 0, 1, false).
			AddItem(originalView, 0, 1, false), 0, 0, 1, 1, 0, 0, false).
		AddItem(tview.NewFlex().
			AddItem(tview.NewBox(), 0, 1, false).
			AddItem(filterInput, 0, 3, true).
			AddItem(tview.NewBox(), 0, 1, false), 1, 0, 1, 1, 0, 0, true)

	elements := []tview.Primitive{outputView, originalView, filterInput}
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
			if e.GetFocusable().HasFocus() {
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

	if err := app.SetRoot(grid, true).SetFocus(grid).Run(); err != nil {
		panic(err)
	}
}
