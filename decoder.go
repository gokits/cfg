package cfg

type Decoder interface {
	Unmarshal(data []byte, v interface{}) error
}

type JsonDecoder int

func (jd *JsonDecoder) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
