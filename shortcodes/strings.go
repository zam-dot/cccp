// strings.go
package shortcodes

import (
	"fmt"
	"strings"
	"text/template"
)

func GetStrings() template.FuncMap {
	return template.FuncMap{

		/* ===================== STRING CREATION ======================= */
		// In strings.go - add this function
		"create_buffer": func(varName string, size int, initialValue string) string {
			// Auto-add quotes if they're missing, but handle the user input properly
			formattedValue := initialValue
			if !strings.HasPrefix(initialValue, `"`) && !strings.HasSuffix(initialValue, `"`) {
				formattedValue = `"` + initialValue + `"`
			}
			return fmt.Sprintf(`char %s[%d] = %s;`, varName, size, formattedValue)
		},
		/* ===================== STRING CREATION ======================= */

		// string_create: Create a new string with automatic memory management
		// Usage: {{ string_create "hello" "str" }}
		// In strings.go - make it handle both quoted and unquoted strings
		"create": func(text, varName string) string {
			// Auto-add quotes if they're missing
			formattedText := text
			if !strings.HasPrefix(text, `"`) && !strings.HasSuffix(text, `"`) {
				formattedText = `"` + text + `"`
			}
			return fmt.Sprintf(`AUTO_FREE char *%s = strdup(%s);`, varName, formattedText)
		},

		/* ===================== STRING CONCATENATION ======================= */

		// Slice: Extract substring from input string
		// Usage: {{ slice "input_var" "start_index" "end_index" "output_var" }}
		// Example: {{ slice "text" "2" "5" "result" }}
		//          // Extracts characters from index 2 to 5 from "text"
		// Requires: input_var must be a valid null-terminated string
		// Output: output_var contains the sliced substring, automatically freed
		"string_concat": func(str1, str2, result string) string {
			return fmt.Sprintf(
				`AUTO_FREE char *%s = malloc(strlen(%s) + strlen(%s) + 1);
if (%s) {
    strcpy(%s, %s);
    strcat(%s, %s);
}`,
				result, str1, str2, result, result, str1, result, str2)
		},

		/* ===================== STRING SLICING ======================= */

		// slice: Python-like string slicing
		// Usage: {{ slice "input_variable" "start_index" "end_index" "output_variable" }}
		"slice": func(source, start, end, result string) string {
			return fmt.Sprintf(
				`AUTO_FREE char *%s = NULL;
if (%s && strlen(%s) >= %s) {
    size_t start_idx = %s;
    size_t end_idx = %s;
    size_t slice_len = end_idx - start_idx;
    %s = malloc(slice_len + 1);
    if (%s) {
        strncpy(%s, %s + start_idx, slice_len);
        %s[slice_len] = '\0';
    }
}`,
				result, source, source, end, start, end, result, result, result, source, result)
		},

		// {{ slice_literal "input_string" "start_index" "end_index" "output_variable" }}
		"slice_literal": func(text, start, end, result string) string {
			return fmt.Sprintf(
				`AUTO_FREE char *%s = NULL;
const char *temp = "%s";
if (strlen(temp) >= %s) {
    size_t slice_len = %s - %s;
    %s = malloc(slice_len + 1);
    if (%s) {
        strncpy(%s, temp + %s, slice_len);
        %s[slice_len] = '\0';
    }
}`,
				result, text, end, end, start, result, result, result, start, result)
		},

		/* ===================== STRING REPEAT ======================= */
		// {{ repeat "input_string" 10 "output_variable" }}
		"repeat": func(text string, count any, result string) string {
			countInt := 1
			switch v := count.(type) {
			case int:
				countInt = v
			case float64:
				countInt = int(v)
			default:
				countInt = 1
			}

			// Generate unique variable names based on result name
			srcVar := result + "_src"
			lenVar := result + "_len"
			countVar := result + "_count"
			iVar := result + "_i"

			return fmt.Sprintf(
				`const char *%s = "%s";
size_t %s = strlen(%s);
size_t %s = %d;
AUTO_FREE char *%s = malloc(%s * %s + 1);
if (%s) {
    %s[0] = '\0';
    for (size_t %s = 0; %s < %s; %s++) {
        strcat(%s, %s);
    }
}`,
				srcVar, text,
				lenVar, srcVar,
				countVar, countInt,
				result, lenVar, countVar,
				result, result,
				iVar, iVar, countVar, iVar,
				result, srcVar)
		},

		/* ===================== STRING FORMATTING ======================= */

		// string_format: Safe sprintf replacement
		// Usage: {{ safe_format "status" "sizeof(status)" "Fetching: %s..." "current_url" }}
		"safe_format": func(buffer, bufferSize, format string, args ...string) string {
			// Build the argument list
			argList := ""
			for i, arg := range args {
				if i > 0 {
					argList += ", "
				}
				argList += arg
			}

			// Ensure the format string has quotes
			formattedFormat := format
			if !strings.HasPrefix(format, `"`) && !strings.HasSuffix(format, `"`) {
				formattedFormat = `"` + format + `"`
			}

			return fmt.Sprintf(
				`if (%s > 0) {
    int written = snprintf(%s, %s, %s, %s);
    if (written >= %s) {
        %s[%s - 1] = '\0';
    }
}`,
				bufferSize, buffer, bufferSize, formattedFormat, argList,
				bufferSize, buffer, bufferSize)
		},

		/* ===================== STRING TRANSFORMS ======================= */

		// string_upper: Convert to uppercase
		// Usage: {{ string_upper "input" "result" }}
		"string_upper": func(input, result string) string {
			return fmt.Sprintf(
				`AUTO_FREE char *%s = %s ? strdup(%s) : NULL;
if (%s) {
    for (char *p = %s; *p; p++) {
        *p = toupper(*p); 
    }
}`,
				result, input, input, result, result)
		},

		// string_lower: Convert to lowercase
		// Usage: {{ string_lower "input" "result" }}
		"string_lower": func(input, result string) string {
			return fmt.Sprintf(
				`AUTO_FREE char *%s = strdup(%s);
if (%s) {
    for (char *p = %s; *p; p++) {
        *p = tolower(*p);
    }
}`,
				result, input, result, result)
		},

		/* ===================== STRING RESULT ======================= */

		// string_result: Get the final string value for use in printf etc.
		// Usage: {{ string_result "variable" }}
		// Example: printf("Result: %s\n", {{ string_result "shouting" }});
		"string_result": func(varName string) string {
			return varName // Just returns the variable name for use in C code
		},

		/* ===================== STRING SEARCHING ======================= */

		// string_find: Find substring position
		// Usage: {{ string_find "haystack" "needle" "position" }}
		"string_find": func(haystack, needle, result string) string {
			return fmt.Sprintf(
				`char *pos = strstr(%s, %s);
%s = pos ? (pos - %s) : -1;`,
				haystack, needle, result, haystack)
		},

		/* ===================== STRING COMPARISON ======================= */

		// string_equals: Safe string comparison
		// Usage: {{ if string_equals "str1" "str2" }} ... {{ end }}
		"string_equals": func(str1, str2 string) string {
			return fmt.Sprintf(`(%s && %s && strcmp(%s, %s) == 0)`,
				str1, str2, str1, str2)
		},
	}
}
