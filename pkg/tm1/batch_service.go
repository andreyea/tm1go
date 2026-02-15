package tm1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/andreyea/tm1go/pkg/models"
)

// BatchService handles TM1 $batch operations.
type BatchService struct {
	rest *RestService
}

// NewBatchService creates a new BatchService instance.
func NewBatchService(rest *RestService) *BatchService {
	return &BatchService{rest: rest}
}

// Batch executes a batch of requests against TM1.
//
// For TM1 versions lower than 12, request URLs in the batch payload are prefixed with /api/v1
// (unless already provided) to match legacy endpoint expectations.
func (bs *BatchService) Batch(ctx context.Context, requests []models.BatchRequest) (*models.BatchResponses, error) {
	adjusted := make([]models.BatchRequest, len(requests))
	copy(adjusted, requests)

	if !IsV1GreaterOrEqualToV2(bs.rest.version, "12.0.0") {
		for i := range adjusted {
			adjusted[i].URL = normalizeBatchRequestURL(adjusted[i].URL, true)
		}
	} else {
		for i := range adjusted {
			adjusted[i].URL = normalizeBatchRequestURL(adjusted[i].URL, false)
		}
	}

	payload := map[string]any{"requests": adjusted}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal batch requests: %w", err)
	}

	resp, err := bs.rest.Post(ctx, "/$batch", bytes.NewReader(payloadJSON))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := &models.BatchResponses{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		body, _ := io.ReadAll(resp.Body)
		if len(body) > 0 {
			return nil, fmt.Errorf("decode batch response: %w: %s", err, string(body))
		}
		return nil, fmt.Errorf("decode batch response: %w", err)
	}

	return result, nil
}

func normalizeBatchRequestURL(requestURL string, addLegacyAPIPrefix bool) string {
	u := strings.TrimSpace(requestURL)
	if u == "" {
		return "/"
	}

	if !strings.HasPrefix(u, "/") {
		u = "/" + u
	}

	if addLegacyAPIPrefix {
		if strings.HasPrefix(u, "/api/v1") {
			return u
		}
		return "/api/v1" + u
	}

	return u
}
