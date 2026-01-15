package main

import (
	"errors"
	"flag"
	"fmt"
)

func runSection(args []string) error {
	fs := flag.NewFlagSet("section", flag.ExitOnError)
	file := fs.String("file", "", "path to .ipynb (defaults to stdin)")
	sets := fs.Int("sets", 1, "number of consecutive sections to return")
	format := fs.String("format", "md", "output format: md, json, or py")
	var queryFlags multiFlag
	fs.Var(&queryFlags, "query", "section query (start:N, match:TEXT, id:ID)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(queryFlags) == 0 {
		return errors.New("section requires at least one --query")
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

	sections, err := collectSections(nb, startIdx, *sets)
	if err != nil {
		return err
	}

	return renderSections(*format, sections)
}

func collectSections(nb *notebook, startIdx, count int) ([]sectionBlock, error) {
	if startIdx < 0 || startIdx >= len(nb.Cells) {
		return nil, fmt.Errorf("start index %d out of range", startIdx)
	}

	idx := startIdx
	var sections []sectionBlock
	var level int
	for len(sections) < count && idx < len(nb.Cells) {
		secStart, secEnd, lvl, err := sectionBounds(nb, idx)
		if err != nil {
			return nil, err
		}
		level = lvl
		sectionCells := cloneCells(nb.Cells[secStart:secEnd])
		sections = append(sections, sectionBlock{Cells: sectionCells})
		nextIdx := findNextPeerHeading(nb, secEnd, level)
		if nextIdx < 0 {
			break
		}
		idx = nextIdx
	}
	return sections, nil
}
