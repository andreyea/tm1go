package tm1go

import (
	"encoding/json"
	"fmt"
)

type View interface {
	GetType() string
	GetName() string
	getBody(bool) (string, error)
}

type ViewWrapper struct {
	View View
}

func (vw *ViewWrapper) UnmarshalJSON(data []byte) error {
	// Step 1: Unmarshal JSON into a generic map to inspect it
	var tmp map[string]interface{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	// Step 2: Determine the type based on some field, e.g., "@odata.type"
	switch tmp["@odata.type"] {
	case "#ibm.tm1.api.v1.NativeView":
		vw.View = &NativeView{}
	case "#ibm.tm1.api.v1.MDXView":
		vw.View = &MDXView{}
	default:
		return fmt.Errorf("unknown view type")
	}

	// Step 3: Unmarshal again into the concrete type
	return json.Unmarshal(data, vw.View)
}
