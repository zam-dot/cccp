package main

import (
	"bytes"
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
	tmpl := template.New("example.c.tmpl").Funcs(shortcodes.GetAllShortcodes())

	// Parse the template file from source/ folder
	tmpl, err := tmpl.ParseFiles("source/example.c.tmpl")
	if err != nil {
		panic(fmt.Sprintf("Failed to parse template: %v", err))
	}

	// Execute template to buffer first
	var generatedContent bytes.Buffer
	err = tmpl.Execute(&generatedContent, nil)
	if err != nil {
		panic(fmt.Sprintf("Failed to execute template: %v", err))
	}

	// Process switch statements
	processedContent := processSwitchStatements(generatedContent.String())

	// ✅ Remove unused imports
	processedContent = removeUnusedImports(processedContent)

	// Write final content to file
	err = os.WriteFile("output/example.c", []byte(processedContent), 0644)
	if err != nil {
		panic(fmt.Sprintf("Failed to create output file: %v", err))
	}

	fmt.Println("✅ Successfully generated output/example.c!")
	formatGeneratedCode("output/example.c")
}

func removeUnusedImports(content string) string {
	// Simple heuristic - remove common unused headers
	// This is basic and might need tuning
	lines := strings.Split(content, "\n")
	var result []string

	for _, line := range lines {
		if strings.Contains(line, "#include") {
			// Keep only if we detect usage of functions from that header
			if shouldKeepInclude(line, content) {
				result = append(result, line)
			}
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

func shouldKeepInclude(includeLine, fullContent string) bool {
	headerToFunctions := map[string][]string{
		"<ctype.h>": {"isalpha", "isdigit", "toupper", "tolower", "isalnum", "isprint"},
		"<string.h>": {
			"strlen",
			"strncpy",
			"strcat",
			"strcmp",
			"strcpy",
			"strncat",
			"memset",
			"memcpy",
		},
		"<stdlib.h>": {"malloc", "free", "atoi", "atof", "calloc", "realloc", "exit", "abort"},
		"<stdio.h>": {
			"printf",
			"fprintf",
			"scanf",
			"fopen",
			"fclose",
			"fgets",
			"fputs",
			"sprintf",
		},
		"<stddef.h>":    {"size_t", "NULL", "offsetof", "ptrdiff_t"},
		"<curl/curl.h>": {"curl_", "CURL", "CURLOPT_"}, // curl functions usually start with curl_
		"<locale.h>":    {"setlocale", "localeconv"},
		"<ncurses.h>":   {"initscr", "printw", "getch", "endwin", "refresh", "move", "addch"},
	}

	for header, functions := range headerToFunctions {
		if strings.Contains(includeLine, header) {
			for _, fn := range functions {
				// More flexible matching for partial function names
				if strings.Contains(fullContent, fn) {
					return true
				}
			}
			//			fmt.Printf("Removing unused header: %s\n", strings.TrimSpace(includeLine))
			return false
		}
	}

	return true // Keep unknown headers to be safe
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
