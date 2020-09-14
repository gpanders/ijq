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

func jq(input string, filter string) (string, error) {
	cmd := exec.Command("jq", "-C", filter)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, input)
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(out), nil

}

func main() {
	var original string
	if len(os.Args) >= 2 {
		bytes, err := ioutil.ReadFile(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}

		original = string(bytes)
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		lines := []string{}
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

		original = strings.Join(lines, "\n")
	}

	app := tview.NewApplication()
	originalView := tview.NewTextView().
		SetDynamicColors(true)

	outputView := tview.NewTextView().
		SetDynamicColors(true).
		SetChangedFunc(func() {
			app.Draw()
		})

	outputWriter := tview.ANSIWriter(outputView)

	orig, err := jq(original, ".")
	if err != nil {
		log.Fatal(err);
	}

	fmt.Fprint(tview.ANSIWriter(originalView), orig)
	fmt.Fprint(outputWriter, orig)

	filterInput := tview.NewInputField().
		SetLabel("Filter: ").
		SetChangedFunc(func(text string) {
			out, err := jq(original, text)
			if err != nil {
				return
			}

			outputView.Clear()
			fmt.Fprint(outputWriter, out)
		}).
		SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyEnter:
				app.Stop()
				fmt.Println(outputView.GetText(true))
			}
		})

	grid := tview.NewGrid().
		SetRows(0, 1).
		SetColumns(0, 0).
		SetBorders(true).
		AddItem(outputView, 0, 0, 1, 1, 0, 0, false).
		AddItem(originalView, 0, 1, 1, 1, 0, 0, false).
		AddItem(filterInput, 1, 0, 1, 2, 50, 0, true)

	if err := app.SetRoot(grid, true).SetFocus(grid).Run(); err != nil {
		panic(err)
	}
}
