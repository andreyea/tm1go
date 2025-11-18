package tm1

import (
	"testing"
)

func TestHTTPError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantMsg    string
	}{
		{
			name:       "404 Not Found",
			statusCode: 404,
			body:       "Resource not found",
			wantMsg:    "tm1 api GET http://example.com/test returned status 404: Resource not found",
		},
		{
			name:       "401 Unauthorized",
			statusCode: 401,
			body:       "Unauthorized",
			wantMsg:    "tm1 api GET http://example.com/test returned status 401: Unauthorized",
		},
		{
			name:       "500 Internal Server Error",
			statusCode: 500,
			body:       "Internal error",
			wantMsg:    "tm1 api GET http://example.com/test returned status 500: Internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &HTTPError{
				StatusCode: tt.statusCode,
				Method:     "GET",
				URL:        "http://example.com/test",
				Body:       []byte(tt.body),
			}

			if err.Error() != tt.wantMsg {
				t.Errorf("HTTPError.Error() = %v, want %v", err.Error(), tt.wantMsg)
			}
		})
	}
}
