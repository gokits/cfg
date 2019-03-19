package cfg

type Decoder interface {
	Unmarshal(data []byte, v interface{}) error
}
