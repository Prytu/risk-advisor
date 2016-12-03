package app

import (
	"encoding/json"
	"fmt"

	"k8s.io/kubernetes/pkg/api/unversioned"
)

// Binding response that will always be returned on successful binding
var bindingResponse []byte

// Initialize 'constant' binding success response
func init() {
	bindingResponseJSON, err := json.Marshal(unversioned.Status{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Status",
			APIVersion: "v1",
		},
		ListMeta: unversioned.ListMeta{},
		Status:   "Success",
		Code:     201,
	})
	if err != nil {
		panic(fmt.Sprintf("Initialization error while marshalling binding response: %v", err))
	}

	bindingResponse = bindingResponseJSON
}
