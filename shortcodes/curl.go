// curl.go - Enhanced version
package shortcodes

import (
	"fmt"
	"strings"
	"text/template"
)

func GetCurl() template.FuncMap {
	return template.FuncMap{
		/* ===================== CURL CLEANUP FUNCTION ======================= */
		"curlCleanupFunc": func() string {
			return `static void curl_cleanup(CURL **curl) {
    if (*curl) {
        curl_easy_cleanup(*curl);
        *curl = NULL;
    }
}`
		},

		/* ==================== FREE HTTP MEMORY ======================= */
		"freeResponse": func(responseVar string) string {
			return fmt.Sprintf(
				`if (%s) { free(%s); %s = NULL; }`,
				responseVar, responseVar, responseVar)
		},

		/* ===================== CURL INIT ======================= */
		"curlInit": func(handle string) string {
			return fmt.Sprintf(
				`CURL *%s = curl_easy_init();
if (!%s) {
    fprintf(stderr, "❌ Failed to initialize CURL\n");
    exit(EXIT_FAILURE);
}
// Auto-cleanup on scope exit
__attribute__((cleanup(curl_cleanup))) CURL *auto_cleanup_%s = %s;`,
				handle, handle, handle, handle)
		},

		/* ===================== CURL SETOPT ======================= */
		"curlSetOpt": func(handle, option, value string) string {
			return fmt.Sprintf(
				`{
    CURLcode res = curl_easy_setopt(%s, %s, %s);
    if (res != CURLE_OK) {
        fprintf(stderr, "❌ curl_easy_setopt failed: %%s (option: %%s)\n", 
                curl_easy_strerror(res), %s);
        exit(EXIT_FAILURE);
    }
}`,
				handle, option, value, option)
		},

		/* ===================== CURL PERFORM ======================= */
		"curlPerform": func(handle string) string {
			return fmt.Sprintf(
				`{
    CURLcode res = curl_easy_perform(%s);
    if (res != CURLE_OK) {
        fprintf(stderr, "❌ HTTP request failed: %%s\n", curl_easy_strerror(res));
        exit(EXIT_FAILURE);
    }
}`,
				handle)
		},

		/* ===================== HTTP CALLBACK ======================= */
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

		/* ===================== HTTP GET ======================= */
		"httpGet": func(url, responseVar string) string {
			formattedURL := url
			if !strings.HasPrefix(url, `"`) && !strings.HasSuffix(url, `"`) {
				formattedURL = `"` + url + `"`
			}

			return fmt.Sprintf(
				`char *%s = NULL;
CURL *curl = curl_easy_init();
if (!curl) {
    fprintf(stderr, "❌ Failed to initialize CURL\n");
    exit(EXIT_FAILURE);
}
// Auto-cleanup on scope exit
__attribute__((cleanup(curl_cleanup))) CURL *auto_cleanup_curl = curl;

// Set options with error checking
{
    CURLcode res = curl_easy_setopt(curl, CURLOPT_URL, %s);
    if (res != CURLE_OK) {
        fprintf(stderr, "❌ curl_easy_setopt failed: %%s (option: CURLOPT_URL)\n", 
                curl_easy_strerror(res));
        exit(EXIT_FAILURE);
    }
}
{
    CURLcode res = curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, write_callback_%s);
    if (res != CURLE_OK) {
        fprintf(stderr, "❌ curl_easy_setopt failed: %%s (option: CURLOPT_WRITEFUNCTION)\n", 
                curl_easy_strerror(res));
        exit(EXIT_FAILURE);
    }
}
{
    CURLcode res = curl_easy_setopt(curl, CURLOPT_WRITEDATA, &%s);
    if (res != CURLE_OK) {
        fprintf(stderr, "❌ curl_easy_setopt failed: %%s (option: CURLOPT_WRITEDATA)\n", 
                curl_easy_strerror(res));
        exit(EXIT_FAILURE);
    }
}
{
    CURLcode res = curl_easy_setopt(curl, CURLOPT_USERAGENT, "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36");
    if (res != CURLE_OK) {
        fprintf(stderr, "❌ curl_easy_setopt failed: %%s (option: CURLOPT_USERAGENT)\n", 
                curl_easy_strerror(res));
        exit(EXIT_FAILURE);
    }
}

// Perform request
{
    CURLcode res = curl_easy_perform(curl);
    if (res != CURLE_OK) {
        fprintf(stderr, "❌ HTTP request failed: %%s\n", curl_easy_strerror(res));
        exit(EXIT_FAILURE);
    }
}

// Check HTTP status code
long http_code = 0;
curl_easy_getinfo(curl, CURLINFO_RESPONSE_CODE, &http_code);
if (http_code != 200) {
    fprintf(stderr, "⚠️  HTTP %%ld: %%s\n", http_code, %s ? %s : "(no response)");
}`,
				responseVar, formattedURL, responseVar, responseVar, responseVar, responseVar)
		},
	}
}
