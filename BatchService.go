package tm1go

import (
	"encoding/json"
)

type BatchService struct {
	rest *RestService
}

func NewBatchService(rest *RestService) *BatchService {
	return &BatchService{rest: rest}
}

// Batch - Execute a batch of requests
// Process body using the following example:
// rsp1 := tm1Go.SomeType{}
// bodyBytes := json.Marshal(response.Responses[i].Body)
// json.Unmarshal(bodyBytes, &rsp1)
func (bs *BatchService) Batch(requests []BatchRequest) (*BatchResponses, error) {
	if !isV1GreaterOrEqualToV2(bs.rest.version, "12.0.0") {
		for i := range requests {
			requests[i].URL = "/api/v1" + requests[i].URL
		}
	}

	requestsAsJSON, err := json.Marshal(requests)
	if err != nil {
		return nil, err
	}
	body := "{\"requests\":" + string(requestsAsJSON) + "}"
	url := "/$batch"
	response, err := bs.rest.POST(url, body, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	batchResponses := &BatchResponses{}
	err = json.NewDecoder(response.Body).Decode(batchResponses)

	if err != nil {
		return nil, err
	}

	return batchResponses, nil
}
