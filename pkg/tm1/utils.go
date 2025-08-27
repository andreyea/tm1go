package tm1

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// TranslateToBool converts various string representations to boolean
func TranslateToBool(value interface{}) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		lower := strings.ToLower(v)
		return lower == "true" || lower == "1" || lower == "yes" || lower == "on"
	case int:
		return v != 0
	case float64:
		return v != 0
	default:
		return false
	}
}

// Base64Decode decodes a base64 encoded string
func Base64Decode(encoded string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// Base64Encode encodes a string to base64
func Base64Encode(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

// CaseAndSpaceInsensitiveEquals compares two strings ignoring case and spaces
func CaseAndSpaceInsensitiveEquals(s1, s2 string) bool {
	clean1 := strings.ReplaceAll(strings.ToLower(s1), " ", "")
	clean2 := strings.ReplaceAll(strings.ToLower(s2), " ", "")
	return clean1 == clean2
}

// ConstructURL builds a URL from components
func ConstructURL(base string, parts ...string) string {
	if !strings.HasSuffix(base, "/") {
		base += "/"
	}

	for i, part := range parts {
		part = strings.Trim(part, "/")
		if i == len(parts)-1 {
			base += part
		} else {
			base += part + "/"
		}
	}

	return base
}

// URLAndBody prepares URL and body for HTTP requests
func URLAndBody(baseURL, path, data string, encoding string) (string, []byte, error) {
	fullURL := ConstructURL(baseURL, path)

	// URL encode if needed
	if u, err := url.Parse(fullURL); err == nil {
		fullURL = u.String()
	}

	// Handle data encoding
	var body []byte
	if data != "" {
		if encoding == "utf-8" || encoding == "" {
			body = []byte(data)
		} else {
			// Handle other encodings if needed
			body = []byte(data)
		}
	}

	return fullURL, body, nil
}

// WaitTimeGenerator generates wait times for exponential backoff
func WaitTimeGenerator(maxTimeout *time.Duration) <-chan time.Duration {
	ch := make(chan time.Duration)

	go func() {
		defer close(ch)
		wait := time.Millisecond * 100 // Start with 100ms
		maxWait := time.Second * 5     // Max 5 seconds between attempts

		start := time.Now()

		for {
			if maxTimeout != nil && time.Since(start) >= *maxTimeout {
				return
			}

			ch <- wait

			// Exponential backoff with jitter
			wait = wait * 2
			if wait > maxWait {
				wait = maxWait
			}
		}
	}()

	return ch
}

// ParseProxies parses proxy configuration from string or map
func ParseProxies(proxies interface{}) (map[string]string, error) {
	if proxies == nil {
		return nil, nil
	}

	switch p := proxies.(type) {
	case map[string]string:
		return p, nil
	case string:
		var result map[string]string
		if err := json.Unmarshal([]byte(p), &result); err != nil {
			return nil, fmt.Errorf("invalid JSON in proxies: %w", err)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("proxies must be map[string]string or JSON string")
	}
}

// ParseTimeout parses timeout from various types
func ParseTimeout(timeout interface{}) (*time.Duration, error) {
	if timeout == nil {
		return nil, nil
	}

	switch t := timeout.(type) {
	case time.Duration:
		return &t, nil
	case float64:
		d := time.Duration(t * float64(time.Second))
		return &d, nil
	case int:
		d := time.Duration(t) * time.Second
		return &d, nil
	case string:
		if f, err := strconv.ParseFloat(t, 64); err == nil {
			d := time.Duration(f * float64(time.Second))
			return &d, nil
		}
		return nil, fmt.Errorf("invalid timeout format: %s", t)
	default:
		return nil, fmt.Errorf("timeout must be duration, float64, int, or string")
	}
}

// ExtractAsyncID extracts async ID from Location header
func ExtractAsyncID(location string) (string, error) {
	// Location format: /api/v1/ExecuteProcessWithReturn('async_id')
	start := strings.Index(location, "('")
	if start == -1 {
		return "", fmt.Errorf("invalid location header format: %s", location)
	}
	start += 2

	end := strings.Index(location[start:], "')")
	if end == -1 {
		return "", fmt.Errorf("invalid location header format: %s", location)
	}

	return location[start : start+end], nil
}

// IsAdmin checks if user is admin (case and space insensitive)
func IsAdmin(username string) bool {
	return CaseAndSpaceInsensitiveEquals(username, "ADMIN")
}

// SanitizeConfigForLogging removes sensitive information from config for logging
func SanitizeConfigForLogging(config map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})
	sensitiveFields := map[string]bool{
		"password":                  true,
		"api_key":                   true,
		"application_client_secret": true,
		"cam_passport":              true,
		"session_id":                true,
	}

	for key, value := range config {
		if sensitiveFields[strings.ToLower(key)] {
			sanitized[key] = "***REDACTED***"
		} else {
			sanitized[key] = value
		}
	}

	return sanitized
}
