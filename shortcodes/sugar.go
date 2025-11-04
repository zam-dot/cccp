package shortcodes

import (
	"fmt"
	"text/template"
)

// GetSugars returns all your syntactic sugar template functions
func GetSugars() template.FuncMap {
	return template.FuncMap{

		// switch: Begin a switch statement (marker for post-processing)
		// Usage: {{ switch "value" }}
		// Example: {{ switch "choice" }}
		"switch": func(value string) string {
			// Return a marker that we'll process later
			return fmt.Sprintf("/* SWITCH_START(%s) */", value)
		},
		// case: Define a case in a switch statement (marker for post-processing)
		// Usage: {{ case "pattern" }}
		// Example: {{ case "1" }}
		"case": func(pattern string) string {
			return fmt.Sprintf("/* CASE(%s) */", pattern)
		},
		// default: Define default case in switch statement (marker for post-processing)
		// Usage: {{ default }}
		// Example: {{ default }}
		"default": func() string {
			return "/* DEFAULT */"
		},
		// end_switch: End a switch statement (marker for post-processing)
		// Usage: {{ end_switch }}
		// Example: {{ end_switch }}
		"end_switch": func() string {
			return "/* SWITCH_END */"
		},
		// Add more sugar here!
	}
}
