// registry.go
package shortcodes

import (
	"fmt"
	"text/template"
)

// Registry of all function providers
var functionProviders = []func() template.FuncMap{
	GetShortcodes, // Core functions (core.go)
	GetCurl,       // HTTP functions (curl.go)
	GetJSON,       // JSON functions (json.go)
	GetSugars,     // Syntactic sugar functions (sugar.go)
	// Add new providers here as you create them
}

// GetAllShortcodes automatically discovers and combines all functions
// Returns: Combined template.FuncMap with all available shortcode functions
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

// ListFunctions returns a list of all available function names
// Usage: Helpful for debugging and discovering available shortcodes
// Returns: Slice of function names as strings
func ListFunctions() []string {
	funcMap := GetAllShortcodes()
	names := make([]string, 0, len(funcMap))
	for name := range funcMap {
		names = append(names, name)
	}
	return names
}
