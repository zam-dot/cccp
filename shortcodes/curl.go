// curl.go
package shortcodes

import (
	"fmt"
	"text/template"
)

func GetCurl() template.FuncMap {
	return template.FuncMap{

		/* ===================== HTTP CALLBACK ======================= */

		// httpCallback: Generate CURL write callback function for HTTP responses
		// Usage: {{ httpCallback "variableName" }} (must be outside main())
		// Example: {{ httpCallback "response" }}
		// Note: Creates a static write_callback_function for handling HTTP response data
		"httpCallback": func(resultVar string) string {
			return fmt.Sprintf(`/* 🌐 HTTP Callback for %s */
static size_t write_callback_%s(char *ptr, size_t size, size_t nmemb, void *userdata) {
    size_t total_size = size * nmemb;
    char **response = (char **)userdata;
    
    size_t current_len = *response ? strlen(*response) : 0;
    char *new_response = realloc(*response, current_len + total_size + 1);
    if (!new_response) {
        fprintf(stderr, "Memory allocation failed in write callback\n");
        return 0;
    }
    
    *response = new_response;
    memcpy(*response + current_len, ptr, total_size);
    (*response)[current_len + total_size] = '\0';
    return total_size;
}`, resultVar, resultVar)
		},

		/* ======================= HTTP GET ========================= */

		// httpGet: Perform HTTP GET request using libcurl
		// Usage: {{ httpGet "url" "responseVariable" }}
		// Example: {{ httpGet "\"https://api.example.com/data\"" "response" }}
		// Note: Requires libcurl development libraries (-lcurl when compiling)
		// Note: Response variable must have corresponding httpCallback declaration
		"httpGet": func(urlVar, resultVar string) string {
			return fmt.Sprintf(`char *%s = NULL;
CURL *curl = curl_easy_init();
if (!curl) {
    fprintf(stderr, "Failed to initialize curl\n");
    exit(EXIT_FAILURE);
}

curl_easy_setopt(curl, CURLOPT_URL, %s);
curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, write_callback_%s);
curl_easy_setopt(curl, CURLOPT_WRITEDATA, &%s);
curl_easy_setopt(curl, CURLOPT_FOLLOWLOCATION, 1L);
curl_easy_setopt(curl, CURLOPT_USERAGENT, "Curl-Sugar/1.0");

CURLcode res = curl_easy_perform(curl);
if (res != CURLE_OK) {
    fprintf(stderr, "HTTP GET failed: %%s\n", curl_easy_strerror(res));
    if (%s) {
        free(%s);
        %s = NULL;
    }
}

curl_easy_cleanup(curl);`, resultVar, urlVar, resultVar, resultVar, resultVar, resultVar, resultVar)
		},

		/* ==================== FREE HTTP MEMORY ======================= */

		// freeResponse: Safely free HTTP response memory
		// Usage: {{ freeResponse "responseVariable" }}
		// Example: {{ freeResponse "response" }}
		// Note: Sets pointer to NULL after freeing to prevent double-free
		"freeResponse": func(responseVar string) string {
			return fmt.Sprintf(
				`if (%s) { free(%s); %s = NULL; }`,
				responseVar,
				responseVar,
				responseVar,
			)
		},
	}
}
