package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type notebook struct {
	Cells []cell `json:"cells"`
}

type cell struct {
	CellType string   `json:"cell_type"`
	Source   nbSource `json:"source"`
	ID       string   `json:"id"`
	Outputs  []output `json:"outputs"`
	Index    int      `json:"_index,omitempty"`
}

type output struct {
	Name       string     `json:"name"`
	OutputType string     `json:"output_type"`
	Text       nbSource   `json:"text"`
	Data       outputData `json:"data"`
	Stream     string     `json:"stream"`
	Ename      string     `json:"ename"`
	Evalue     string     `json:"evalue"`
	Traceback  []string   `json:"traceback"`
	Metadata   outputMeta `json:"metadata"`
}

type outputData map[string]any

type outputMeta map[string]any

type nbSource []string

func (s *nbSource) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		*s = nil
		return nil
	}
	switch data[0] {
	case '[':
		var arr []string
		if err := json.Unmarshal(data, &arr); err != nil {
			return err
		}
		*s = nbSource(arr)
	case '"':
		var str string
		if err := json.Unmarshal(data, &str); err != nil {
			return err
		}
		*s = nbSource([]string{str})
	default:
		return fmt.Errorf("unsupported source encoding: %s", string(data))
	}
	return nil
}

type heading struct {
	Level   int    `json:"level"`
	Title   string `json:"title"`
	Preview string `json:"preview"`
}

func readNotebook(path string) (*notebook, error) {
	var data []byte
	var err error
	if path != "" {
		data, err = os.ReadFile(path)
	} else {
		data, err = io.ReadAll(os.Stdin)
	}
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("no notebook data")
	}

	var nb notebook
	if err := json.Unmarshal(data, &nb); err != nil {
		return nil, err
	}
	for i := range nb.Cells {
		nb.Cells[i].Index = i
	}
	return &nb, nil
}

func collectHeadings(nb *notebook, previewWords int) []heading {
	var result []heading
	for _, c := range nb.Cells {
		if c.CellType != "markdown" {
			continue
		}
		result = append(result, headingsFromCell(c, previewWords)...)
	}
	return result
}

func headingsFromCell(c cell, previewWords int) []heading {
	var hs []heading
	lines := c.Source
	for idx, raw := range lines {
		trimmed := strings.TrimLeft(raw, " ")
		if !strings.HasPrefix(trimmed, "#") {
			continue
		}
		level := countLeadingHashes(trimmed)
		if level == 0 {
			continue
		}
		title := strings.TrimSpace(trimmed[level:])
		preview := previewFromLines(lines[idx+1:], previewWords)
		hs = append(hs, heading{Level: level, Title: title, Preview: preview})
	}
	return hs
}

func countLeadingHashes(s string) int {
	count := 0
	for _, r := range s {
		if r == '#' {
			count++
			continue
		}
		break
	}
	return count
}

func previewFromLines(lines []string, limit int) string {
	if limit <= 0 {
		return ""
	}
	var words []string
	truncated := false
	for _, line := range lines {
		for _, w := range strings.Fields(line) {
			words = append(words, w)
			if len(words) == limit {
				truncated = true
				break
			}
		}
		if truncated {
			break
		}
	}
	if len(words) == 0 {
		return ""
	}
	preview := strings.Join(words, " ")
	if truncated {
		preview += " ..."
	}
	return preview
}

func firstHeadingLevel(c cell) (int, bool) {
	if c.CellType != "markdown" {
		return 0, false
	}
	for _, raw := range c.Source {
		trimmed := strings.TrimLeft(raw, " ")
		if !strings.HasPrefix(trimmed, "#") {
			continue
		}
		level := countLeadingHashes(trimmed)
		if level == 0 {
			continue
		}
		return level, true
	}
	return 0, false
}

func sectionBounds(nb *notebook, startIdx int) (int, int, int, error) {
	cell := nb.Cells[startIdx]
	level, ok := firstHeadingLevel(cell)
	if !ok {
		return 0, 0, 0, fmt.Errorf("cell %d is not a markdown heading", startIdx)
	}
	end := len(nb.Cells)
	for i := startIdx + 1; i < len(nb.Cells); i++ {
		lvl, ok := firstHeadingLevel(nb.Cells[i])
		if ok && lvl <= level {
			end = i
			break
		}
	}
	return startIdx, end, level, nil
}

func findNextPeerHeading(nb *notebook, start int, level int) int {
	for i := start; i < len(nb.Cells); i++ {
		lvl, ok := firstHeadingLevel(nb.Cells[i])
		if !ok {
			continue
		}
		if lvl == level {
			return i
		}
	}
	return -1
}

func cellText(c cell) string {
	return strings.Join(c.Source, "")
}

func cloneCells(cells []cell) []cell {
	res := make([]cell, len(cells))
	copy(res, cells)
	return res
}

func excludeOutputs(cells []cell) []cell {
	res := make([]cell, len(cells))
	for i, c := range cells {
		res[i] = c
		res[i].Outputs = nil
	}
	return res
}
