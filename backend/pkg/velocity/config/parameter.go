package config

import "encoding/json"

type parameter interface{}

type baseParameter struct {
	Type string `json:"type"`
}

func NewParameterBasic() *parameterBasic {
	return &parameterBasic{
		baseParameter: baseParameter{
			Type: "basic",
		},
	}
}

type parameterBasic struct {
	baseParameter
	Name         string   `json:"name"`
	Default      string   `json:"default"`
	OtherOptions []string `json:"otherOptions"`
	Secret       bool     `json:"secret"`
}

func NewParameterDerived() *parameterDerived {
	return &parameterDerived{
		baseParameter: baseParameter{
			Type: "derived",
		},
	}
}

type parameterDerived struct {
	baseParameter
	Use       string            `json:"use"`
	Secret    bool              `json:"secret"`
	Arguments map[string]string `json:"arguments"`
	Exports   map[string]string `json:"exports"`
	Timeout   uint64            `json:"timeout"`
}

func unmarshalParameter(b []byte) (p parameter, err error) {
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
