package shortcodes

import (
	"fmt"
	"text/template"
)

// GetShortcodes returns all your C template functions
func GetShortcodes() template.FuncMap {
	return template.FuncMap{

		/* ===================== READ DYNAMIC LINES ======================= */

		// readDynamicLinePortable: Read a line from file into dynamically allocated buffer
		// Usage: {{ readDynamicLinePortable "filePointer" "lineBuffer" }}
		// Example: {{ readDynamicLinePortable "fp" "line" }}
		"readDynamicLinePortable": func(fpVar, lineVar string) string {
			return fmt.Sprintf(
				`char *%s = NULL;
size_t buffer_size = 1024;
char *buffer = malloc(buffer_size);
if (buffer == NULL) {
    fprintf(stderr, "Buffer allocation failed\n");
    exit(EXIT_FAILURE);
}

// Use getline for automatic buffer growth if available
// Or read in chunks for portability
if (fgets(buffer, buffer_size, %s) != NULL) {
    size_t len = strlen(buffer);
    %s = malloc(len + 1);
    if (%s == NULL) {
        fprintf(stderr, "Memory allocation failed\n");
        free(buffer);
        exit(EXIT_FAILURE);
    }
    strncpy(%s, buffer, len);
    %s[len] = '\0';
    // Remove trailing newline if present
    if (len > 0 && %s[len-1] == '\n') {
        %s[len-1] = '\0';
    }
} else {
    if (feof(%s)) {
        %s = NULL;
    } else {
        fprintf(stderr, "Error reading from file\n");
        free(buffer);
        exit(EXIT_FAILURE);
    }
}
free(buffer);`,
				lineVar,
				fpVar,
				lineVar,
				lineVar,
				lineVar,
				lineVar,
				lineVar,
				lineVar,
				fpVar,
				lineVar,
			)
		},

		/* ======================= STRING COPY =========================== */

		// stringCopy: Safe string copy with bounds checking
		// Usage: {{ stringCopy "destination" "source" "bufferSize" }}
		// Example: {{ stringCopy "dest" "src" "sizeof(dest)" }}
		"stringCopy": func(dest, src, destSize string) string {
			return fmt.Sprintf(
				`if (%s > 0) {
    strncpy(%s, %s, %s - 1);
    %s[%s - 1] = '\\0';
} else {
    fprintf(stderr, "Invalid buffer size for string copy\\n");
}`,
				destSize, dest, src, destSize, dest, destSize)
		},

		/* ==================== STRING CONCATENATION ======================= */

		// strcat: Safe string concatenation with bounds checking
		// Usage: {{ strcat "destination" "source" "bufferSize" }}
		// Example: {{ strcat "dest" "src" "sizeof(dest)" }}
		"strcat": func(dest, src, destSize string) string {
			return fmt.Sprintf(
				`// Safe strcat with bounds checking
size_t dest_len = strlen(%s);
size_t src_len = strlen(%s);
if (dest_len + src_len < %s - 1) {
    strncpy(%s + dest_len, %s, %s - dest_len - 1);
    %s[dest_len + src_len] = '\\0';
} else {
    fprintf(stderr, "strcat would overflow buffer\\n");
    // Let the caller decide how to handle the error
}`,
				dest, src, destSize, dest, src, destSize, dest)
		},

		/* ======================= OPEN FILE ============================ */

		// openFile: Open a file with error handling
		// Usage: {{ openFile "filename" "mode" "filePointerVar" }}
		// Example: {{ openFile "\"data.txt\"" "\"r\"" "fp" }}
		"openFile": func(filename, mode, varName string) string {
			return fmt.Sprintf(
				`FILE *%s = fopen("%s", "%s");
if (%s == NULL) {
    fprintf(stderr, "ERROR: Cannot open '%%s' in mode '%%s'\n", "%s", "%s");
    perror("fopen failed");
    exit(EXIT_FAILURE);
}
// ⚠️ Remember to fclose(%s)!`,
				varName, filename, mode, varName, filename, mode, varName)
		},

		/* ===================== Auto Free Generic ======================== */

		// autoFreeGeneric: Define automatic memory cleanup macro (GCC/Clang)
		// Usage: Include once in your template, no parameters needed
		// Example: {{ autoFreeGeneric }}
		"autoFreeGeneric": func() string {
			return `#if defined(__GNUC__) || defined(__clang__)
    #define AUTO_FREE __attribute__((cleanup(auto_free_generic)))
#else
    #define AUTO_FREE // No support for other compilers
    // Manual cleanup required for non-GCC/Clang
#endif

static void auto_free_generic(void *p) { 
    free(*(void**)p); 
}`
		},

		/* ======================= GET MEMORY ========================== */

		// getMemory: Allocate memory with automatic cleanup (GCC/Clang only)
		// Usage: {{ getMemory "type" "variable" count }}
		// Example: {{ getMemory "int" "numbers" 10 }}
		"getMemory": func(typeName, varName string, count int) string {
			return fmt.Sprintf(
				`AUTO_FREE %s *%s = malloc(%d * sizeof(%s));
if (%s == NULL) {
    fprintf(stderr, "Memory allocation failed for %s\\n");
    exit(EXIT_FAILURE);
}`,
				typeName, varName, count, typeName, varName, varName)
		},

		/* ======================= GROW MEMORY ========================== */

		// growMemory: Reallocate memory to grow an existing buffer
		// Usage: {{ growMemory "pointerName" newCount }}
		// Example: {{ growMemory "buffer" 200 }}
		"growMemory": func(ptrName string, newCount int) string {
			return fmt.Sprintf(
				`%s = realloc(%s, %d * sizeof(*%s));
if (%s == NULL) {
    fprintf(stderr, "Memory reallocation failed for %s\n");
    exit(1);
}`,
				ptrName, ptrName, newCount, ptrName, ptrName, ptrName)
		},

		/* ======================= CREATE ARRAY  ========================== */

		// createArray: Create a dynamic array with auto-growing capability
		// Usage: {{ createArray "type" "variable" initialSize }}
		// Example: {{ createArray "int" "myArray" 10 }}
		"createArray": func(typeName, varName string, initialSize int) string {
			return fmt.Sprintf(
				`// Auto-growing array
typedef struct {
    %s *data;
    size_t size;
    size_t capacity;
} Array_%s;

Array_%s %s = {
    .data = malloc(%d * sizeof(%s)),
    .size = 0,
    .capacity = %d
};

if (%s.data == NULL) {
    fprintf(stderr, "Array allocation failed for %s\n");
    exit(EXIT_FAILURE);
}`,
				typeName, varName, varName, varName, initialSize, typeName,
				initialSize, varName, varName)
		},

		/* ======================== PUSH ARRAY  =========================== */

		// push: Add element to dynamic array, auto-growing if needed
		// Usage: {{ push "arrayName" "value" }}
		// Example: {{ push "myArray" "42" }}
		"push": func(arrayName, value string) string {
			return fmt.Sprintf(
				`// Auto-grow array if needed
if (%s.size >= %s.capacity) {
    %s.capacity *= 2;
    %s.data = realloc(%s.data, %s.capacity * sizeof(*%s.data));
    if (%s.data == NULL) {
        fprintf(stderr, "Array growth failed for %s\n");
        exit(EXIT_FAILURE);
    }
    printf("Array grown to capacity: %%zu\n", %s.capacity);
}
%s.data[%s.size++] = %s;`,
				arrayName, arrayName, arrayName, arrayName, arrayName,
				arrayName, arrayName, arrayName, arrayName, arrayName,
				arrayName, arrayName, value)
		},

		/* ======================== ARRAY CLEANUP  ============================ */

		// arrayCleanup: Generate automatic cleanup function for dynamic arrays
		// Usage: {{ arrayCleanup "arrayVariable" }}
		// Example: {{ arrayCleanup "myArray" }}
		"arrayCleanup": func(varName string) string {
			return fmt.Sprintf(
				`static void auto_free_array_%s(void *p) { 
    Array_%s *arr = p;
    if (arr->data != NULL) {
        free(arr->data);
        arr->data = NULL;
    }
}
__attribute__((cleanup(auto_free_array_%s))) Array_%s %s;`,
				varName, varName, varName, varName, varName)
		},

		/* ========================== READ LINE  ============================= */

		// readLine: Read a line into fixed-size buffer with newline removal
		// Usage: {{ readLine "filePointer" "buffer" "bufferSize" }}
		// Example: {{ readLine "fp" "lineBuffer" "sizeof(lineBuffer)" }}
		"readLine": func(fpVar, bufferVar, bufferSize string) string {
			return fmt.Sprintf(
				`if (fgets(%s, %s, %s) == NULL) {
    if (feof(%s)) {
        // Handle EOF
        %s[0] = '\0';
    } else {
        fprintf(stderr, "Error reading from file\n");
        exit(EXIT_FAILURE);
    }
}
// Remove trailing newline if present
size_t len = strlen(%s);
if (len > 0 && %s[len-1] == '\n') {
    %s[len-1] = '\0';
}`,
				bufferVar, bufferSize, fpVar, fpVar, bufferVar, bufferVar, bufferVar, bufferVar)
		},

		/* ========================= CREATE STRING ============================ */

		// createString: Create a string builder for efficient string construction
		// Usage: {{ createString "variable" }}
		// Example: {{ createString "builder" }}
		"createString": func(varName string) string {
			return fmt.Sprintf(
				`// Auto-growing string builder
typedef struct {
    char *data;
    size_t length;
    size_t capacity;
} StringBuilder_%s;

StringBuilder_%s %s = {
    .data = malloc(16 * sizeof(char)),
    .length = 0,
    .capacity = 16
};

if (%s.data == NULL) {
    fprintf(stderr, "String builder allocation failed for %s\n");
    exit(EXIT_FAILURE);
}
%s.data[0] = '\0';`, // Fixed: single backslash
				varName, varName, varName, varName, varName, varName)
		},

		/* ========================== APPEND STRING ============================= */

		// append: Append text to string builder
		// Usage: {{ append "builderName" "text" }}
		// Example: {{ append "builder" "\"Hello, World!\"" }}
		"append": func(builderName, text string) string {
			// Generate unique variable names using builderName
			return fmt.Sprintf(
				`// Append to string builder
{
    size_t append_len_%s = strlen(%s);
    if (%s.length + append_len_%s + 1 > %s.capacity) {
        // Double capacity until it fits
        while (%s.length + append_len_%s + 1 > %s.capacity) {
            %s.capacity *= 2;
        }
        %s.data = realloc(%s.data, %s.capacity);
        if (%s.data == NULL) {
            fprintf(stderr, "String builder growth failed for %s\n");
            exit(EXIT_FAILURE);
        }
    }
    strcpy(%s.data + %s.length, %s);
    %s.length += append_len_%s;
}`,
				builderName, text,
				builderName, builderName, builderName,
				builderName, builderName, builderName,
				builderName, builderName, builderName, builderName,
				builderName, builderName,
				builderName, builderName, text,
				builderName, builderName)
		},

		/* ========================= STRING RESULT  ============================ */

		// stringResult: Get the final string from string builder
		// Usage: {{ stringResult "builderName" }}
		// Example: printf("Result: %s\n", {{ stringResult "builder" }});
		"stringResult": func(builderName string) string {
			return fmt.Sprintf("%s.data", builderName)
		},
	}
}
