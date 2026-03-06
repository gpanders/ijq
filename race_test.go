//go:build race

package main

import (
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"codeberg.org/gpanders/ijq/internal/options"
)

func TestAppRace(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("race reproduction test relies on a shell helper script")
	}

	// Keep the test self-contained by disabling history writes and using the
	// helper jq wrapper that keeps processing active while events are spammed.
	cfg := DefaultConfig()
	cfg.HistoryFile = ""
	cfg.JQCommand = "./testdata/catok"

	// Use enough input data to keep filter evaluation and rendering busy during
	// concurrent key events.
	doc := Document{
		input:  strings.Repeat(`{"foo":1,"bar":2,"baz":3}`+"\n", 20),
		filter: ".",
		options: options.Options{
			HistoryFile: cfg.HistoryFile,
			JQCommand:   cfg.JQCommand,
		},
		config: cfg,
	}

	// Run the app on a simulation screen so we can deterministically inject key
	// events without requiring a real terminal.
	app := createApp(doc)
	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatalf("init simulation screen: %v", err)
	}

	app.SetScreen(screen)

	// app.Run blocks until the app exits, so run it in the background and collect
	// any returned error through a channel.
	runErr := make(chan error, 1)
	go func() {
		runErr <- app.Run()
	}()

	// Give the event loop a short moment to initialize before flooding events.
	time.Sleep(50 * time.Millisecond)

	// Prime the UI with an initial navigation key event.
	screen.PostEventWait(tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModShift))

	var wg sync.WaitGroup
	const iterations = 100

	wg.Go(func() {
		// Repeatedly type a filter expression to stress the filter-edit path.
		for i := 0; i < iterations; i++ {
			for _, r := range ".foo.bar" {
				screen.PostEventWait(tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone))
			}
		}
	})

	wg.Go(func() {
		// Trigger a command key in parallel to race with filter updates.
		for i := 0; i < iterations; i++ {
			screen.PostEventWait(tcell.NewEventKey(tcell.KeyCtrlO, ' ', tcell.ModNone))
		}
	})

	wg.Wait()

	app.Stop()

	select {
	case err := <-runErr:
		if err != nil {
			t.Fatalf("run app: %v", err)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("timed out waiting for app to stop")
	}
}
