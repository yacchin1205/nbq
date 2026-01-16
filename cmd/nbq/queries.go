package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const queryUsage = "start:N, match:REGEX, contains:TEXT, id:ID"

type multiFlag []string

func (m *multiFlag) String() string {
	return strings.Join(*m, ",")
}

func (m *multiFlag) Set(value string) error {
	*m = append(*m, value)
	return nil
}

type queryFilter struct {
	start    *int
	match    []*regexp.Regexp
	contains []string
	id       *string
}

func parseQueryFlags(values []string) (queryFilter, error) {
	var filter queryFilter
	for _, raw := range values {
		parts := strings.SplitN(raw, ":", 2)
		if len(parts) != 2 {
			return filter, fmt.Errorf("invalid query: %s", raw)
		}
		key := parts[0]
		value := parts[1]
		switch key {
		case "start":
			if filter.start != nil {
				return filter, errors.New("start query specified multiple times")
			}
			idx, err := strconv.Atoi(value)
			if err != nil {
				return filter, fmt.Errorf("invalid start index: %s", value)
			}
			filter.start = &idx
		case "match":
			re, err := regexp.Compile(value)
			if err != nil {
				return filter, fmt.Errorf("invalid regex in match: %s", err)
			}
			filter.match = append(filter.match, re)
		case "contains":
			filter.contains = append(filter.contains, value)
		case "id":
			if filter.id != nil {
				return filter, errors.New("id query specified multiple times")
			}
			filter.id = &value
		default:
			return filter, fmt.Errorf("unsupported query key: %s", key)
		}
	}
	return filter, nil
}

func locateStartCell(nb *notebook, filter queryFilter) (int, error) {
	if filter.start != nil {
		idx := *filter.start
		if idx < 0 || idx >= len(nb.Cells) {
			return 0, fmt.Errorf("start index %d out of range", idx)
		}
		if !matchesCell(nb.Cells[idx], idx, filter) {
			return 0, fmt.Errorf("cell %d does not satisfy remaining queries", idx)
		}
		return idx, nil
	}
	for idx, c := range nb.Cells {
		if matchesCell(c, idx, filter) {
			return idx, nil
		}
	}
	return 0, errors.New("no cell matched the query")
}

func matchesCell(c cell, idx int, filter queryFilter) bool {
	if filter.id != nil {
		if c.ID == "" || c.ID != *filter.id {
			return false
		}
	}
	text := cellText(c)
	for _, re := range filter.match {
		if !re.MatchString(text) {
			return false
		}
	}
	for _, s := range filter.contains {
		if !strings.Contains(text, s) {
			return false
		}
	}
	return true
}
