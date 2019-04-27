package builder

import (
	"fmt"
	"os"
)

type ParameterResolver struct {
	Params map[string]string
}

func NewParameterResolver(params map[string]string) *ParameterResolver {
	return &ParameterResolver{
		Params: params,
	}
}

func (pR *ParameterResolver) Resolve(paramName string) (string, error) {

	if val, ok := pR.Params[paramName]; ok {
		return val, nil
	}

	fromEnv := os.Getenv(fmt.Sprintf("VCI_%s", paramName))
	if len(fromEnv) > 0 {
		return fromEnv, nil
	}

	return "", fmt.Errorf("parameter %s not defined", paramName)
}
