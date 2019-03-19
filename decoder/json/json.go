package json

import (
	js "encoding/json"
)

type JsonDecoder int

func (jd *JsonDecoder) Unmarshal(data []byte, v interface{}) error {
	return js.Unmarshal(data, v)
}
