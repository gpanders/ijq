package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

var ErrEmptyPath = errors.New("no path specified")

type history struct {
	Path string
}

func (h *history) Add(expression string) error {
	if expression == "" {
		return nil
	}

	if h.Path == "" {
		return ErrEmptyPath
	}

	saved, err := h.Get()
	if err != nil {
		return err
	}

	// Don't cotinue with adding the expression if
	// it is saved in history already.
	if contains(saved, expression) {
		return nil
	}

	file, err := h.openFile()
	if err != nil {
		return fmt.Errorf(
			"error opening history for writing: %w", err,
		)
	}

	fmt.Fprintln(file, expression)

	if err = file.Close(); err != nil {
		return fmt.Errorf(
			"error closing history file: %w", err,
		)
	}

	return nil
}

func (h *history) Get() ([]string, error) {
	var expressions []string

	filebytes, err := ioutil.ReadFile(h.Path)
	if err != nil {
		// If the history file doesn't exist, then
		// return an empty history.
		if errors.Is(err, os.ErrNotExist) {
			return expressions, nil
		} else {
			return nil, fmt.Errorf(
				"error retrieving history: %w", err,
			)
		}
	}

	scanner := bufio.NewScanner(bytes.NewReader(filebytes))
	for scanner.Scan() {
		expressions = append(expressions, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf(
			"error retrieving history: %w", err,
		)
	}

	return expressions, nil
}

func (h *history) openFile() (*os.File, error) {
	err := os.MkdirAll(filepath.Dir(h.Path), os.ModePerm)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(
		h.Path,
		os.O_WRONLY|os.O_CREATE|os.O_APPEND,
		0644,
	)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func contains(arr []string, elem string) bool {
	for _, v := range arr {
		if elem == v {
			return true
		}
	}

	return false
}
