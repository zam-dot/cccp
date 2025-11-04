// json.go
package shortcodes

import (
	"fmt"
	"text/template"
)

func GetJSON() template.FuncMap {
	return template.FuncMap{
		"jsonExtract": func(jsonVar, pathVar, resultVar string) string {
			return fmt.Sprintf(`/* 📄 Extract JSON value using jq */
char *%s = NULL;
{
    // Write JSON to temporary file to avoid shell escaping issues
    FILE *tmp = fopen("temp_json.json", "w");
    if (!tmp) {
        fprintf(stderr, "Failed to create temp file\n");
        exit(EXIT_FAILURE);
    }
    fprintf(tmp, "%%s", %s);
    fclose(tmp);
    
    // Use jq on the temp file - path is already quoted in the format string
    char command[256];
    snprintf(command, sizeof(command), "jq -r %%s temp_json.json", %s);
    
    FILE *pipe = popen(command, "r");
    if (!pipe) {
        fprintf(stderr, "Failed to run jq\n");
        exit(EXIT_FAILURE);
    }
    
    %s = malloc(1024);
    if (fgets(%s, 1024, pipe) == NULL) {
        fprintf(stderr, "Failed to read from jq\n");
        free(%s);
        %s = NULL;
    }
    pclose(pipe);
    
    // Remove temp file
    remove("temp_json.json");
    
    // Remove trailing newline
    if (%s) {
        size_t len = strlen(%s);
        if (len > 0 && %s[len-1] == '\n') {
            %s[len-1] = '\0';
        }
    }
}`, resultVar, jsonVar, pathVar, resultVar, resultVar, resultVar, resultVar, resultVar, resultVar, resultVar, resultVar)
		},
	}
}
