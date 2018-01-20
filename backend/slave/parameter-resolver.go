package main

import "fmt"

type ParameterResolver struct {
	Params map[string]string
}

func NewParameterResolver(params map[string]string) ParameterResolver {
	return ParameterResolver{
		Params: params,
	}
}

func (pR *ParameterResolver) Resolve(paramName string) (string, error) {

	// TODO: check env?

	if val, ok := pR.Params[paramName]; ok {
		return val, nil
	}

	return "", fmt.Errorf("parameter %s not defined", paramName)
}
