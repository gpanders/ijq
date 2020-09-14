package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type Document struct {
	contents string
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
	cmd := exec.Command("jq", "-C", filter)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, d.contents)
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(out), nil

}

func main() {
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

	var doc Document
	go func() {
		if len(os.Args) > 1 {
			if err := doc.FromFile(os.Args[1]); err != nil {
				log.Fatal(err)
			}
		} else {
			if err := doc.FromStdin(); err != nil {
				log.Fatal(err)
			}
		}

		out, err := doc.Filter(".")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Fprint(tview.ANSIWriter(originalView), out)
		fmt.Fprint(outputWriter, out)
	}()

	filterInput := tview.NewInputField().
		SetFieldBackgroundColor(0).
		SetFieldTextColor(7).
		SetChangedFunc(func(text string) {
			go func() {
				outputView.Clear()
				out, err := doc.Filter(text)
				if err != nil {
					fmt.Fprintf(outputWriter, "Invalid filter")
					return
				}

				fmt.Fprint(outputWriter, out)
				outputView.ScrollToBeginning()
			}()
		}).
		SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyEnter:
				app.Stop()
				fmt.Println(outputView.GetText(true))
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
			AddItem(tview.NewBox(), 0, 1, false), 1, 0, 1, 1, 50, 0, true)

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
