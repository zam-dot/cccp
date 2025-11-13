package generators

import (
	"fmt"
	"strings"

	"github.com/flosch/pongo2/v6"
)

func init() {
	Register(InitFileFilters)
}

func InitFileFilters() {
	// Safe file open with error checking
	// Example usage:
	// FILE *config_file;
	// {{ "config_file" | safe_fopen : "config.txt,r" }}
	pongo2.RegisterFilter("safe_fopen", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		fileVar := in.String()
		params := strings.Split(param.String(), ",")
		if len(params) != 2 {
			return nil, &pongo2.Error{OrigError: fmt.Errorf("safe_fopen needs filename,mode")}
		}

		code := fmt.Sprintf(
			`%[1]s = fopen("%[2]s", "%[3]s");
if (!%[1]s) {
    fprintf(stderr, "Failed to open file: %s\n", "%[2]s");
    exit(EXIT_FAILURE);
}`,
			fileVar, params[0], params[1])
		return pongo2.AsSafeValue(code), nil
	})
	// Example usage:
	// DIR *dir;
	// {{ "dir" | open_directory : "path" }}
	pongo2.RegisterFilter("open_directory", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		dirVar := in.String()
		path := param.String()

		code := fmt.Sprintf(
			`%[1]s = opendir(%[2]s);
if (!%[1]s) {
    fprintf(stderr, "Failed to open directory: %%s\n", %[2]s);
    exit(EXIT_FAILURE);
}`,
			dirVar, path)
		return pongo2.AsSafeValue(code), nil
	})
	// Example usage:
	// {{ "dir" | close_directory }}
	pongo2.RegisterFilter("close_directory", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		dirVar := in.String()

		code := fmt.Sprintf(
			`if (%[1]s) {
    closedir(%[1]s);
    %[1]s = NULL;
}`,
			dirVar)
		return pongo2.AsSafeValue(code), nil
	})
}
