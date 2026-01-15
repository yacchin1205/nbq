package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func printHeadingsMarkdown(headings []heading) {
	for _, h := range headings {
		fmt.Printf("%s %s\n", strings.Repeat("#", h.Level), h.Title)
		fmt.Println()
		if h.Preview != "" {
			fmt.Printf("%s\n", h.Preview)
		}
		fmt.Println()
	}
}

type sectionBlock struct {
	Cells []cell
}

func renderSections(format string, sections []sectionBlock) error {
	switch format {
	case "md":
		printSectionsMarkdown(sections)
	case "json":
		cells := flattenSectionCells(sections)
		payload := map[string]any{
			"cells": cells,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(payload); err != nil {
			return err
		}
	case "py":
		printSectionsPython(sections)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
	return nil
}

func printSectionsMarkdown(sections []sectionBlock) {
	for idx, section := range sections {
		if idx > 0 {
			fmt.Println("---")
		}
		for _, c := range section.Cells {
			switch c.CellType {
			case "markdown":
				text := cellText(c)
				fmt.Print(text)
				if !strings.HasSuffix(text, "\n") {
					fmt.Println()
				}
				fmt.Println()
			case "code":
				fmt.Println("```")
				code := cellText(c)
				fmt.Print(code)
				if !strings.HasSuffix(code, "\n") {
					fmt.Println()
				}
				fmt.Println("```")
				fmt.Println()
			default:
				fmt.Print(cellText(c))
				fmt.Println()
			}
		}
	}
}

func printSectionsPython(sections []sectionBlock) {
	for idx, section := range sections {
		if idx > 0 {
			fmt.Println("# ---")
		}
		for _, c := range section.Cells {
			switch c.CellType {
			case "markdown":
				writeMarkdownAsComments(c)
			case "code":
				code := cellText(c)
				fmt.Print(code)
				if !strings.HasSuffix(code, "\n") {
					fmt.Println()
				}
				fmt.Println()
			default:
				fmt.Printf("# %s\n\n", cellText(c))
			}
		}
	}
}

func writeMarkdownAsComments(c cell) {
	for _, line := range c.Source {
		line = strings.TrimRight(line, "\n")
		if strings.TrimSpace(line) == "" {
			fmt.Println("#")
			continue
		}
		fmt.Printf("# %s\n", line)
	}
	fmt.Println()
}

func flattenSectionCells(sections []sectionBlock) []cell {
	var cells []cell
	for _, section := range sections {
		cells = append(cells, section.Cells...)
	}
	return cells
}
