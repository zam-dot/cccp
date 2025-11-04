// registry.go
package shortcodes

import (
	"fmt"
	"text/template"
)

// Registry of all function providers
var functionProviders = []func() template.FuncMap{
	GetShortcodes, // Core functions
	GetCurl,       // HTTP functions
	GetJSON,       // JSON functions  ← ADD THIS LINE
	// Add new providers here as you create them
}

// GetAllShortcodes automatically discovers and combines all functions
func GetAllShortcodes() template.FuncMap {
	combined := template.FuncMap{}

	for _, provider := range functionProviders {
		funcMap := provider()
		for name, function := range funcMap {
			if _, exists := combined[name]; exists {
				// Handle naming conflicts gracefully
				fmt.Printf("Warning: Function '%s' already exists, skipping\n", name)
				continue
			}
			combined[name] = function
		}
	}

	return combined
}

// Helper to see what functions are available
func ListFunctions() []string {
	funcMap := GetAllShortcodes()
	names := make([]string, 0, len(funcMap))
	for name := range funcMap {
		names = append(names, name)
	}
	return names
}

