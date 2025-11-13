package main

import (
	"fmt"
	"os"
	"os/exec"

	"ctranspiler/pkg/generators"

	"github.com/flosch/pongo2/v6"
)

func main() {
	// Initialize all generators
	generators.InitAll()

	if err := runGeneration(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	formatGeneratedCode("output/main.c")
}

func formatGeneratedCode(filename string) error {
	cmd := exec.Command("clang-format", "-i", filename)
	return cmd.Run()
}

func runGeneration() error {
	// File operations only - read template, write output
	tpl, err := pongo2.FromFile("src/main.c.tpl")
	if err != nil {
		return err
	}

	output, err := tpl.Execute(pongo2.Context{})
	if err != nil {
		return err
	}

	return os.WriteFile("output/main.c", []byte(output), 0o644)
}
