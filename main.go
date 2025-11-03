package main

import (
	"fmt"
	"os"
	"os/exec"
	"template/shortcodes"
	"text/template"
)

func main() {
	// Get all shortcodes from the other file
	funcMap := shortcodes.GetShortcodes()

	// Create output folder
	os.MkdirAll("output", 0755)

	// Create output file
	outputFile, err := os.Create("output/generated.c")
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	// Parse and execute template
	tmpl, err := template.New("test.c.tmpl").Funcs(funcMap).ParseFiles("source/test.c.tmpl")
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(outputFile, nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("✅ Successfully generated output/generated.c!")
	formatGeneratedCode("output/generated.c")
}

func formatGeneratedCode(filename string) error {
	cmd := exec.Command("clang-format", "-i", filename)
	return cmd.Run()
}
