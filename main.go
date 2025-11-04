package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"template/shortcodes"
	"text/template"
)

func main() {
	// One function to rule them all!
	tmpl := template.New("example.c.tmpl").Funcs(shortcodes.GetAllShortcodes())

	// See what's available
	fmt.Println("Available functions:", shortcodes.ListFunctions())

	// ✅ FIX: Parse the template file from source/ folder
	tmpl, err := tmpl.ParseFiles("source/example.c.tmpl")
	if err != nil {
		panic(fmt.Sprintf("Failed to parse template: %v", err))
	}

	// ✅ ADD THIS: Create output file
	outputFile, err := os.Create("output/generated.c")
	if err != nil {
		panic(fmt.Sprintf("Failed to create output file: %v", err))
	}
	defer outputFile.Close()

	// ✅ ADD THIS: Execute the template (this generates the content)
	err = tmpl.Execute(outputFile, nil) // or pass data if needed
	if err != nil {
		panic(fmt.Sprintf("Failed to execute template: %v", err))
	}

	// Now process the generated content
	generatedContent, err := os.ReadFile("output/generated.c")
	if err != nil {
		panic(err)
	}

	fmt.Println("Curl functions:", shortcodes.GetCurl())

	processedContent := processSwitchStatements(string(generatedContent))

	err = os.WriteFile("output/generated.c", []byte(processedContent), 0644)
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

func processSwitchStatements(content string) string {
	// Use regex for exact pattern matching
	content = regexp.MustCompile(`/\* SWITCH_START\(([^)]+)\) \*/`).
		ReplaceAllString(content, "switch ($1) {")
	content = regexp.MustCompile(`/\* CASE\(([^)]+)\) \*/`).
		ReplaceAllString(content, "    case $1:")
	content = strings.ReplaceAll(content, "/* DEFAULT */", "    default:")
	content = strings.ReplaceAll(content, "/* SWITCH_END */", "}")

	return content
}
