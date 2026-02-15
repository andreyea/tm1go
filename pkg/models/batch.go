package models

// BatchResponses represents the payload returned by TM1 $batch endpoint.
type BatchResponses struct {
	Responses []BatchResponse `json:"responses"`
}

// BatchResponse captures one response entry from a TM1 batch call.
type BatchResponse struct {
	ID      string            `json:"id"`
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    interface{}       `json:"body,omitempty"`
}

// BatchRequest captures one request entry for the TM1 $batch endpoint.
type BatchRequest struct {
	Method    string            `json:"method"`
	URL       string            `json:"url"`
	ID        string            `json:"id"`
	Body      interface{}       `json:"body,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
	DependsOn []string          `json:"dependsOn,omitempty"`
}
