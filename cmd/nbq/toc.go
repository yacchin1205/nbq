package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

func runTOC(args []string) error {
	fs := flag.NewFlagSet("toc", flag.ExitOnError)
	file := fs.String("file", "", "path to .ipynb (defaults to stdin)")
	words := fs.Int("words", 20, "number of preview words")
	format := fs.String("format", "md", "output format: md or json")
	if err := fs.Parse(args); err != nil {
		return err
	}

	nb, err := readNotebook(*file)
	if err != nil {
		return err
	}

	headings := collectHeadings(nb, *words)
	if len(headings) == 0 {
		return nil
	}

	switch *format {
	case "md":
		printHeadingsMarkdown(headings)
	case "json":
		headingCells := extractHeadingCells(nb)
		payload := map[string]any{
			"cells": headingCells,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(payload); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported format: %s", *format)
	}

	return nil
}

func extractHeadingCells(nb *notebook) []cell {
	var cells []cell
	for _, c := range nb.Cells {
		if c.CellType != "markdown" {
			continue
		}
		for _, line := range c.Source {
			if strings.HasPrefix(strings.TrimLeft(line, " "), "#") {
				cells = append(cells, c)
				break
			}
		}
	}
	return cells
}
