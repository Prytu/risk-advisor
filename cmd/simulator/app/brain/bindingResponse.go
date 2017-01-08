package brain

import (
	"encoding/json"
	"fmt"

	"k8s.io/client-go/pkg/apis/meta/v1"
)

// TODO: Generate real binding response

// Binding response that will always be returned on successful binding
var bindingResponse []byte

// Initialize 'constant' binding success response
func init() {
	bindingResponseJSON, err := json.Marshal(v1.Status{
		TypeMeta: v1.TypeMeta{
			Kind:       "Status",
			APIVersion: "v1",
		},
		ListMeta: v1.ListMeta{},
		Status:   "Success",
		Code:     201,
	})
	if err != nil {
		panic(fmt.Sprintf("Initialization error while marshalling binding response: %v", err))
	}

	bindingResponse = bindingResponseJSON
}
