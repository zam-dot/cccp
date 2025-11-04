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

		// string_create: Create a new string with automatic memory management
		// Usage: {{ string_create "\"hello\"" "str" }}
		// In strings.go - make it handle both quoted and unquoted strings
		"string_create": func(value, varName string) string {
			// Auto-add quotes if they're missing
			formattedValue := value
			if !strings.HasPrefix(value, `"`) && !strings.HasSuffix(value, `"`) {
				formattedValue = `"` + value + `"`
			}

			return fmt.Sprintf(`AUTO_FREE char *%s = strdup(%s);`, varName, formattedValue)
		},

		/* ===================== STRING CONCATENATION ======================= */

		// string_concat: Safe string concatenation
		// Usage: {{ string_concat "str1" "str2" "result" }}
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

		// string_slice: Python-like string slicing
		// Usage: {{ string_slice "source" "2" "5" "result" }}
		"string_slice": func(source, start, end, result string) string {
			return fmt.Sprintf(
				`AUTO_FREE char *%s = malloc(%s - %s + 1);
if (%s && %s && %s + %s <= strlen(%s)) {
    strncpy(%s, %s + %s, %s - %s);
    %s[%s - %s] = '\0';
} else {
    %s = NULL;
}`,
				result, end, start, result, source, source, start, source,
				result, source, start, end, start, result, end, start, result)
		},

		/* ===================== STRING REPEAT ======================= */

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
		// Usage: {{ string_format "result" "\"Hello %%s, you have %%d messages\"" "name" "count" }}
		"string_format": func(result, format string, args ...string) string {
			argList := ""
			for i, arg := range args {
				if i > 0 {
					argList += ", "
				}
				argList += arg
			}

			return fmt.Sprintf(
				`{
    size_t needed = snprintf(NULL, 0, %s, %s) + 1;
    AUTO_FREE char *%s = malloc(needed);
    if (%s) {
        snprintf(%s, needed, %s, %s);
    }
}`,
				format, argList, result, result, result, format, argList)
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
