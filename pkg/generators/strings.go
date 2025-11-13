package generators

import (
	"fmt"
	"strings"

	"github.com/flosch/pongo2/v6"
)

func init() {
	Register(InitStringFilters)
}

func InitStringFilters() {
	// Example usage:
	// char src[] = "Hello World";
	// char dest[20]; // Must be declared first!
	//
	// strncpy(dest, src, sizeof(dest) - 1);
	// dest[sizeof(dest) - 1] = '\0';
	//
	// printf("Source: %s\n", src);
	// printf("Copy: %s\n", dest);
	pongo2.RegisterFilter("string_copy", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		dest := in.String()
		src := param.String()
		code := fmt.Sprintf("strncpy(%[1]s, %[2]s, sizeof(%[1]s) - 1);\n%[1]s[sizeof(%[1]s) - 1] = '\\0';",
			dest, src)
		return pongo2.AsSafeValue(code), nil
	})

	// Example usage: --> Needs {{ "" | auto_free_generic }}
	// const char* original_name = "Hello World";
	// {{ "uppercase_copy" | string_upper_copy : "original_name" }}
	// printf("%s\n", uppercase_copy);
	pongo2.RegisterFilter("string_upper_copy", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		dest := in.String()
		src := param.String()
		code := fmt.Sprintf(
			`// copy a string and make it uppercase
AUTO_FREE char *%[1]s = %[2]s ? strdup(%[2]s) : NULL;
if (%[1]s) {
    size_t len = strlen(%[1]s);
    for (size_t i = 0; i < len; i++) {  // Explicit length check
        %[1]s[i] = toupper((unsigned char)%[1]s[i]);
    }
    %[1]s[len] = '\0';  // Ensure null termination
}`,
			dest, src) // This line was missing the closing parenthesis
		return pongo2.AsSafeValue(code), nil
	})

	// Example usage:
	// {{ "Sensor reading: " | write_string }}
	// {{ "42" | write_string }}
	// {{ " units" | write_string }}
	// Only provide write_string for optimal output
	pongo2.RegisterFilter("write_string", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		str := in.String()
		return pongo2.AsSafeValue(fmt.Sprintf(`write(1, "%s", %d);`, str, len(str))), nil
	})

	// {{ "" | newline }}
	// Maybe one for newlines since it's common
	pongo2.RegisterFilter("newline", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		return pongo2.AsSafeValue(`write(1, "\n", 1);`), nil
	})

	// Safe string copy with bounds checking
	// Example usage:
	// char path[256];
	// {{ "path" | string_copy : "some_string" }}
	pongo2.RegisterFilter("string_copy", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		dest := in.String()
		src := param.String()

		code := fmt.Sprintf(
			`strncpy(%[1]s, %[2]s, sizeof(%[1]s) - 1);
%[1]s[sizeof(%[1]s) - 1] = '\0';`,
			dest, src)
		return pongo2.AsSafeValue(code), nil
	})

	// Example usage:
	// {{ "" | snprintf_checked : "playlist[track_count],needed,\"%s/\",entry->d_name" }}
	pongo2.RegisterFilter("snprintf_checked", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		// This one needs multiple parameters, so we'll handle it differently
		// Let's assume param contains "dest,size,format,args..."
		parts := strings.Split(param.String(), ",")
		if len(parts) < 3 {
			return nil, &pongo2.Error{OrigError: fmt.Errorf("snprintf_checked needs dest,size,format[,args...]")}
		}

		dest := parts[0]
		size := parts[1]
		format := parts[2]
		args := ""
		if len(parts) > 3 {
			args = "," + strings.Join(parts[3:], ",")
		}

		code := fmt.Sprintf(
			`int _written = snprintf(%[1]s, %[2]s, %[3]s%[4]s);
if (_written < 0 || _written >= (int)%[2]s) {
    fprintf(stderr, "String truncation detected in %%s\n", __func__);
}`,
			dest, size, format, args)
		return pongo2.AsSafeValue(code), nil
	})
}
