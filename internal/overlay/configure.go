package overlay

import (
	"fmt"
	"reflect"

	"codeberg.org/gpanders/ijq/internal/options"
)

// ConfigureRows returns the rows to display in the Configure window
func ConfigureRows(opts options.Options) []string {
	var rows []string
	opts.Each(func(opt options.Option) {
		value := reflect.ValueOf(opt).Elem()
		if value.Kind() == reflect.Bool {
			checkbox := "○"
			if value.Bool() {
				checkbox = "●"
			}

			rows = append(rows, fmt.Sprintf("%s %s (-%s)", checkbox, opt, opt.Flag()))
		}
	})

	return rows
}

var configureRows = func() []options.Option {
	var o options.Options

	var rows []options.Option
	o.Each(func(opt options.Option) {
		value := reflect.Indirect(reflect.ValueOf(opt))
		if !value.IsValid() {
			return
		}

		if value.Kind() != reflect.Bool {
			return
		}

		rows = append(rows, opt)
	})

	return rows
}()

var configureSize = func() struct {
	Width  int
	Height int
} {
	contentWidth := len("Configure")
	for _, row := range ConfigureRows(options.Options{}) {
		contentWidth = max(contentWidth, len(row))
	}

	return struct {
		Width  int
		Height int
	}{
		Width:  contentWidth,
		Height: len(configureRows) + 1,
	}
}()
