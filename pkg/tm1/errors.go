package tm1

import (
	"fmt"
	"io"
	"net/http"
)

// HTTPError captures non-success responses returned by TM1 REST API.
type HTTPError struct {
	Method     string
	URL        string
	StatusCode int
	Body       []byte
}

func (e *HTTPError) Error() string {
	if len(e.Body) == 0 {
		return fmt.Sprintf("tm1 api %s %s returned status %d", e.Method, e.URL, e.StatusCode)
	}
	return fmt.Sprintf("tm1 api %s %s returned status %d: %s", e.Method, e.URL, e.StatusCode, string(e.Body))
}

func newHTTPError(resp *http.Response) error {
	if resp.StatusCode < http.StatusBadRequest {
		return nil
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	return &HTTPError{
		Method:     resp.Request.Method,
		URL:        resp.Request.URL.String(),
		StatusCode: resp.StatusCode,
		Body:       body,
	}
}
