package generators

import (
	"github.com/flosch/pongo2/v6"
)

func init() {
	Register(InitErrorFilters)
}

func InitErrorFilters() {
	// Generate error checking macros
	// Example usage:
	// {{ "" | generate_error_macros }}
	// Then in code:
	// CHECK_NULL(buffer, "audio buffer");
	// CHECK_SYS_CALL(write(fd, data, size), "write failed");
	pongo2.RegisterFilter("generate_error_macros", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		code := `
#define CHECK_NULL(ptr, msg) do { \
    if (!(ptr)) { \
        fprintf(stderr, "NULL pointer: %s\n", msg); \
        exit(EXIT_FAILURE); \
    } \
} while(0)

#define CHECK_SYS_CALL(result, msg) do { \
    if ((result) == -1) { \
        perror(msg); \
        exit(EXIT_FAILURE); \
    } \
} while(0)`

		return pongo2.AsSafeValue(code), nil
	})
}
