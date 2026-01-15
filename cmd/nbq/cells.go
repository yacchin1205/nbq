package main

import (
	"errors"
	"flag"
	"fmt"
)

func runCells(args []string) error {
	fs := flag.NewFlagSet("cells", flag.ExitOnError)
	file := fs.String("file", "", "path to .ipynb (defaults to stdin)")
	sets := fs.Int("sets", 1, "number of Markdown+code pairs")
	format := fs.String("format", "md", "output format: md, json, or py")
	var queryFlags multiFlag
	fs.Var(&queryFlags, "query", "cell query (start:N, match:TEXT, id:ID)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(queryFlags) == 0 {
		return errors.New("cells requires at least one --query")
	}
	if *sets <= 0 {
		return errors.New("--sets must be >= 1")
	}

	nb, err := readNotebook(*file)
	if err != nil {
		return err
	}

	filter, err := parseQueryFlags(queryFlags)
	if err != nil {
		return err
	}

	startIdx, err := locateStartCell(nb, filter)
	if err != nil {
		return err
	}

	sections, err := collectCellSets(nb, startIdx, *sets)
	if err != nil {
		return err
	}

	return renderSections(*format, sections)
}

func collectCellSets(nb *notebook, startIdx, count int) ([]sectionBlock, error) {
	if startIdx < 0 || startIdx >= len(nb.Cells) {
		return nil, fmt.Errorf("start index %d out of range", startIdx)
	}
	idx := startIdx
	var sections []sectionBlock
	for len(sections) < count {
		if idx >= len(nb.Cells) {
			return nil, errors.New("not enough Markdown+code sets")
		}
		if nb.Cells[idx].CellType != "markdown" {
			return nil, fmt.Errorf("cell %d is not a markdown cell", idx)
		}
		end := idx + 1
		for end < len(nb.Cells) && nb.Cells[end].CellType == "code" {
			end++
		}
		sections = append(sections, sectionBlock{Cells: cloneCells(nb.Cells[idx:end])})
		idx = end
	}
	return sections, nil
}
