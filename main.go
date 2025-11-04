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
	// Create template with all available shortcode functions
	tmpl := template.New("strings.c.tmpl").Funcs(shortcodes.GetAllShortcodes())

	// Parse the template file from source/ folder
	tmpl, err := tmpl.ParseFiles("source/strings.c.tmpl")
	if err != nil {
		panic(fmt.Sprintf("Failed to parse template: %v", err))
	}

	// Create output file for generated C code
	outputFile, err := os.Create("output/strings.c")
	if err != nil {
		panic(fmt.Sprintf("Failed to create output file: %v", err))
	}
	defer outputFile.Close()

	// Execute the template to generate C code content
	err = tmpl.Execute(outputFile, nil)
	if err != nil {
		panic(fmt.Sprintf("Failed to execute template: %v", err))
	}

	// Process the generated content for syntactic sugar
	generatedContent, err := os.ReadFile("output/strings.c")
	if err != nil {
		panic(err)
	}

	// Process switch statements (convert markers to actual C switch syntax)
	processedContent := processSwitchStatements(string(generatedContent))

	// Write the processed content back to the file
	err = os.WriteFile("output/strings.c", []byte(processedContent), 0644)
	if err != nil {
		panic(err)
	}

	fmt.Println("✅ Successfully generated output/strings.c!")

	// Format the generated code for better readability
	formatGeneratedCode("output/strings.c")
}

// formatGeneratedCode runs clang-format on the generated C file
// Requires: clang-format to be installed on the system
// Usage: Automatically formats the generated code to follow C style guidelines
func formatGeneratedCode(filename string) error {
	cmd := exec.Command("clang-format", "-i", filename)
	return cmd.Run()
}

// processSwitchStatements converts syntactic sugar markers to actual C switch syntax
// Processes: SWITCH_START, CASE, DEFAULT, and SWITCH_END markers
// Usage: Post-processes template output to generate valid C switch statements
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
