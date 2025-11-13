package generators

import (
	"fmt"
	"strings"

	"github.com/flosch/pongo2/v6"
)

func init() {
	Register(InitMemoryFilters)
}

func InitMemoryFilters() {
	// Example usage:
	// {{ "" | auto_free_generic }}  // Include once at top of file
	//
	// Then in functions:
	// AUTO_FREE char* buffer = malloc(100);  // Automatically freed!
	//
	// Note: Only works on GCC/Clang, falls back to no-op on other compilers
	pongo2.RegisterFilter("auto_free_generic", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		code := `#if defined(__GNUC__) || defined(__clang__)
#define AUTO_FREE __attribute__((cleanup(auto_free_generic)))
#else
    #define AUTO_FREE
#endif

static void auto_free_generic(void *p) { 
    free(*(void**)p); 
}`
		return pongo2.AsSafeValue(code), nil
	})

	// Generates safe malloc with error checking
	// Example usage:
	// {{ "buffer" | get_memory : "1024" }}
	pongo2.RegisterFilter("get_memory", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		dest := in.String()
		size := param.String()
		code := fmt.Sprintf(
			`%[1]s = malloc(%[2]s);
if (!%[1]s) {
    fprintf(stderr, "Failed to get memory for %[1]s (size: %%zu)\n", (size_t)%[2]s);
    exit(EXIT_FAILURE);
}`,
			dest, size)
		return pongo2.AsSafeValue(code), nil
	})

	// Extend your AUTO_FREE to handle files, DIR*, etc
	// {{ "" | generate_auto_cleanup }}
	// Now use anywhere:
	// AUTO_FREE char *buffer = malloc(100);
	// AUTO_FILE FILE *logfile = fopen("log.txt", "w");
	// AUTO_DIR DIR *dir = opendir("/path");
	pongo2.RegisterFilter("generate_auto_cleanup", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		code := `#include <stdlib.h>  // for free
#include <stdio.h>   // for FILE, fclose  
#include <dirent.h>  // for DIR, closedir

#if defined(__GNUC__) || defined(__clang__)
// Generic cleanup for any type
#define AUTO_CLEANUP(cleanup_func) __attribute__((cleanup(cleanup_func)))

// Specific cleaners
static void auto_free_generic(void *p) { free(*(void**)p); }
static void auto_close_file(void *p) { if (*(FILE**)p) fclose(*(FILE**)p); }
static void auto_close_dir(void *p) { if (*(DIR**)p) closedir(*(DIR**)p); }

// Convenience macros
#define AUTO_FREE AUTO_CLEANUP(auto_free_generic)
#define AUTO_FILE AUTO_CLEANUP(auto_close_file)  
#define AUTO_DIR AUTO_CLEANUP(auto_close_dir)
#else
#define AUTO_FREE
#define AUTO_FILE
#define AUTO_DIR
#endif`

		return pongo2.AsSafeValue(code), nil
	})
	// Example usage:
	// {{ "playlist[track_count]" | copy_string : "\"../\"" }}
	pongo2.RegisterFilter("copy_string", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		dest := in.String()
		src := param.String()

		code := fmt.Sprintf(
			`strncpy(%[1]s, %[2]s, sizeof(%[1]s) - 1);
%[1]s[sizeof(%[1]s) - 1] = '\0';`,
			dest, src)
		return pongo2.AsSafeValue(code), nil
	})

	// Example usage:
	// struct Config *config;
	// {{ "config" | get_zeroed_memory : "sizeof(struct Config)" }}
	// config is now all zeros instead of garbage values

	// char *buffer;
	// {{ "buffer" | get_zeroed_memory : "1024" }}
	// buffer is now all zeros instead of uninitialized
	pongo2.RegisterFilter("get_zeroed_memory", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		dest := in.String()
		size := param.String()

		code := fmt.Sprintf(
			`%[1]s = calloc(1, %[2]s);
if (!%[1]s) {
    fprintf(stderr, "Failed to get zeroed memory for %[1]s (size: %%zu)\n", (size_t)%[2]s);
    exit(EXIT_FAILURE);
}`,
			dest, size)
		return pongo2.AsSafeValue(code), nil
	})

	// Example usage:
	// {{ "playlist" | auto_cleanup_array : "track_count" }}
	pongo2.RegisterFilter("auto_cleanup_array", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		arrayVar := in.String()
		countVar := param.String()

		code := fmt.Sprintf(
			`for (int i = 0; i < %[2]s; i++) {
    if (%[1]s[i]) {
        free(%[1]s[i]);
        %[1]s[i] = NULL;
    }
}
%[2]s = 0;`,
			arrayVar, countVar)
		return pongo2.AsSafeValue(code), nil
	})

	// Example usage:
	// FILE* config = load_config();
	// {{ "config" | check_null : "config loading" }}
	// char* input = get_user_input();
	// {{ "input" | check_null : "user input" }}
	// {{ "buffer" | check_null : "buffer validation" }}
	pongo2.RegisterFilter("check_null", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		ptr := in.String()
		context := param.String()
		code := fmt.Sprintf(
			`if (%[1]s == NULL) { 
    fprintf(stderr, "NULL pointer in %%s: %[2]s\n", __func__); 
    exit(EXIT_FAILURE); 
}`,
			ptr, context)
		return pongo2.AsSafeValue(code), nil
	})

	// Example usage:
	// int fd = {{ "open(\"data.txt\", O_RDONLY)" | check_syscall : "file opening" }};
	//
	// Network operations
	// int sockfd = {{ "socket(AF_INET, SOCK_STREAM, 0)" | check_syscall : "socket creation" }};
	// Process operations
	// {{ "fork()" | check_syscall : "process forking" }}
	pongo2.RegisterFilter("check_syscall", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		call := in.String()
		context := param.String()
		code := fmt.Sprintf(
			`if (%[1]s == -1) { 
    perror("System call failed in %[2]s"); 
    exit(EXIT_FAILURE); 
}`,
			call, context)
		return pongo2.AsSafeValue(code), nil
	})

	// Example usage:
	// for (int i = 0; i < count; i++) {
	//      {{ "i,array_size" | check_bounds }}
	//      process_item(array[i]);
	// }
	pongo2.RegisterFilter("check_bounds", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		parts := strings.Split(in.String(), ",")
		if len(parts) != 2 {
			return nil, &pongo2.Error{OrigError: fmt.Errorf("check_bounds needs index,size")}
		}
		index, size := parts[0], parts[1]
		code := fmt.Sprintf(
			`if (%[1]s >= %[2]s) { 
    fprintf(stderr, "Index %%zu out of bounds (size: %%zu) in %%s\n", (size_t)%[1]s, (size_t)%[2]s, __func__); 
    exit(EXIT_FAILURE); 
}`,
			index, size)
		return pongo2.AsSafeValue(code), nil
	})

	// Example usage:
	// {{ "" | generate_error_macros }}
	pongo2.RegisterFilter("generate_error_macros", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		code := `#include <stdio.h>
#include <stdlib.h>

#define CHECK_NULL(ptr, msg) do { \
    if (!(ptr)) { \
        fprintf(stderr, "NULL pointer: %s in %s\n", msg, __func__); \
        exit(EXIT_FAILURE); \
    } \
} while (0)

#define CHECK_SYS_CALL(result, msg) do { \
    if ((result) == -1) { \
        perror(msg); \
        exit(EXIT_FAILURE); \
    } \
} while(0)

#define CHECK_BOUNDS(index, size, msg) do { \
    if ((index) >= (size)) { \
        fprintf(stderr, "Bounds check failed: %s (index: %%zu, size: %%zu) in %%s\n", \
                msg, (size_t)(index), (size_t)(size), __func__); \
        exit(EXIT_FAILURE); \
    } \
} while(0)`

		return pongo2.AsSafeValue(code), nil
	})

	// Add this to your error handling package

	pongo2.RegisterFilter("check_args", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		condition := in.String()
		message := param.String()
		code := fmt.Sprintf(
			`if (%[1]s) { 
    fprintf(stderr, "Invalid arguments: %[2]s\n"); 
    fprintf(stderr, "Usage: %%s <source> <dest>\n", argv[0]); 
    exit(EXIT_FAILURE); 
}`,
			condition, message)
		return pongo2.AsSafeValue(code), nil
	})

	// For the read/write size validation, use this:
	pongo2.RegisterFilter("check_min_size", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		parts := strings.Split(in.String(), ",")
		if len(parts) != 2 {
			return nil, &pongo2.Error{OrigError: fmt.Errorf("check_min_size needs actual,expected")}
		}
		actual, expected := parts[0], parts[1]
		code := fmt.Sprintf(
			`if (%[1]s < %[2]s) { 
    fprintf(stderr, "Size check failed: got %%zd, expected at least %%zd in %%s\n", 
            (size_t)%[1]s, (size_t)%[2]s, __func__); 
    exit(EXIT_FAILURE); 
}`,
			actual, expected)
		return pongo2.AsSafeValue(code), nil
	})
}
