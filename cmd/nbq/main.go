package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	case "toc":
		err = runTOC(args)
	case "section":
		err = runSection(args)
	case "cells":
		err = runCells(args)
	case "outputs":
		err = runOutputs(args)
	case "-h", "--help", "help":
		usage()
		return
	default:
		err = fmt.Errorf("unknown command: %s", cmd)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "nbq: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "nbq <command> [options]\ncommands: toc, section, cells, outputs")
}
