package tm1go

type BatchResponses struct {
	Responses []BatchResponse `json:"responses"`
}

type BatchResponse struct {
	ID      string            `json:"id"`
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    interface{}       `json:"body,omitempty"`
}

type BatchRequest struct {
	Method    string            `json:"method"`
	URL       string            `json:"url"`
	ID        string            `json:"id"`
	Body      string            `json:"body,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
	DependsOn []string          `json:"dependsOn,omitempty"`
}
