package yaml

import yaml "gopkg.in/yaml.v3"

type YamlDecoder int

func (yd *YamlDecoder) Unmarshal(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}
