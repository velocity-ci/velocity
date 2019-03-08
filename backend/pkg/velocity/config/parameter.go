package config

import "encoding/json"

type Parameter interface {
}

type BaseParameter struct {
	Type string `json:"type"`
}

func NewParameterBasic() *ParameterBasic {
	return &ParameterBasic{
		BaseParameter: BaseParameter{
			Type: "basic",
		},
	}
}

type ParameterBasic struct {
	BaseParameter
	Name         string   `json:"name"`
	Default      string   `json:"default"`
	OtherOptions []string `json:"otherOptions"`
	Secret       bool     `json:"secret"`
}

func NewParameterDerived() *ParameterDerived {
	return &ParameterDerived{
		BaseParameter: BaseParameter{
			Type: "derived",
		},
	}
}

type ParameterDerived struct {
	BaseParameter
	Use       string            `json:"use"`
	Secret    bool              `json:"secret"`
	Arguments map[string]string `json:"arguments"`
	Exports   map[string]string `json:"exports"`
	Timeout   uint64            `json:"timeout"`
}

func unmarshalParameter(b []byte) (p Parameter, err error) {
	var m map[string]interface{}
	err = json.Unmarshal(b, &m)
	if err != nil {
		return p, err
	}

	if _, ok := m["use"]; ok { // derived
		p = NewParameterDerived()
	} else if _, ok := m["name"]; ok { // basic
		p = NewParameterBasic()
	}

	err = json.Unmarshal(b, p)

	return p, err
}
