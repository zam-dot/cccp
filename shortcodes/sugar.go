package shortcodes

import (
	"fmt"
	"text/template"
)

// GetSugars returns all your syntactic sugar template functions
func GetSugars() template.FuncMap {
	return template.FuncMap{

		// In sugar.go
		"switch": func(value string) string {
			// Return a marker that we'll process later
			return fmt.Sprintf("/* SWITCH_START(%s) */", value)
		},

		"case": func(pattern string) string {
			return fmt.Sprintf("/* CASE(%s) */", pattern)
		},

		"default": func() string {
			return "/* DEFAULT */"
		},

		"end_switch": func() string {
			return "/* SWITCH_END */"
		},
		// Add more sugar here!
	}
}
