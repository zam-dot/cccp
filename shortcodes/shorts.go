package shortcodes

import (
	"fmt"
	"text/template"
)

// GetShortcodes returns all your C template functions
func GetShortcodes() template.FuncMap {
	return template.FuncMap{
		// ====== Dynamic Line ======
		// Add this to your shorts.go
		"readDynamicLinePortable": func(fpVar, lineVar string) string {
			return fmt.Sprintf(
				`char *%s = NULL;
char buffer[1024];
if (fgets(buffer, sizeof(buffer), %s) != NULL) {
    size_t len = strlen(buffer);
    %s = malloc(len + 1);
    if (%s == NULL) {
        fprintf(stderr, "Memory allocation failed\n");
        exit(EXIT_FAILURE);
    }
    strcpy(%s, buffer);
    // Remove trailing newline if present
    if (len > 0 && %s[len-1] == '\n') {
        %s[len-1] = '\0';
    }
} else {
    if (feof(%s)) {
        %s = NULL;
    } else {
        fprintf(stderr, "Error reading from file\n");
        exit(EXIT_FAILURE);
    }
}`,
				lineVar, // 1 - declare line
				fpVar,   // 2 - file pointer
				lineVar, // 3 - malloc
				lineVar, // 4 - check if NULL
				lineVar, // 5 - strcpy dest
				lineVar, // 6 - check for newline
				lineVar, // 7 - remove newline
				fpVar,   // 8 - feof check
				lineVar, // 9 - set to NULL
			)
		},

		// ====== Read Line =======
		"readLine": func(fpVar, bufferVar string) string {
			return fmt.Sprintf(
				`if (fgets(%s, sizeof(%s), %s) == NULL) {
    if (feof(%s)) {
        // Handle EOF
    } else {
        fprintf(stderr, "Error reading from file\n");
        exit(EXIT_FAILURE);
    }
}
// Remove trailing newline if present
%s[strcspn(%s, "\n")] = '\0';`,
				bufferVar, bufferVar, fpVar, fpVar, bufferVar, bufferVar)
		},

		// ====== String Copy ======
		"stringCopy": func(dest, src, maxSize string) string {
			return fmt.Sprintf(
				`strncpy(%s, %s, %s - 1);
%s[%s - 1] = '\0';`,
				dest, src, maxSize, dest, maxSize)
		},

		// ====== String Copy ======
		"safeStrcat": func(dest, src string) string {
			return fmt.Sprintf(
				`// Safe strcat with bounds checking
size_t dest_len = strlen(%s);
size_t src_len = strlen(%s);
if (dest_len + src_len < 50) {  // Use the actual buffer size
    strcpy(%s + dest_len, %s);
} else {
    fprintf(stderr, "strcat would overflow buffer\n");
    // Let the caller decide how to handle the error
}`,
				dest, src, dest, src)
		},

		// ====== Open File ======
		"openFile": func(filename, mode, varName string) string {
			return fmt.Sprintf(
				`FILE *%s = fopen("%s", "%s");
if (%s == NULL) {
    fprintf(stderr, "ERROR: Cannot open '%s' in mode '%s'\n");
    perror("fopen failed");
    exit(EXIT_FAILURE);
}
// ⚠️ Remember to fclose(%s)!`,
				varName, filename, mode, varName, filename, mode, varName)
		},

		// ====== Get Memory ======
		// Add this once at file level
		"autoFreeGeneric": func() string {
			return `static void auto_free_generic(void *p) { 
    free(*(void**)p); 
}`
		},

		// Updated getMemory with auto-free
		"getMemory": func(typeName, varName string, count int) string {
			return fmt.Sprintf(
				`__attribute__((cleanup(auto_free_generic))) %s *%s = malloc(%d * sizeof(%s));
if (%s == NULL) {
    fprintf(stderr, "Memory allocation failed for %s\n");
    exit(EXIT_FAILURE);
}`,
				typeName, varName, count, typeName, varName, varName)
		},

		// ====== Grow Memory ======
		"growMemory": func(ptrName string, newCount int) string {
			return fmt.Sprintf(
				`%s = realloc(%s, %d * sizeof(*%s));
if (%s == NULL) {
    fprintf(stderr, "Memory reallocation failed for %s\n");
    exit(1);
}`,
				ptrName, ptrName, newCount, ptrName, ptrName, ptrName)
		},
		// Add more shortcodes here as you create them!
	}
}
