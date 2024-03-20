package requester

import (
	"encoding/json"
)

type Marshaller interface {
	Marshal(value any) ([]byte, error)
}

type JSONMarshaller struct{}

func (jm *JSONMarshaller) Marshal(value any) ([]byte, error) {
	return json.Marshal(value)
}
