package v3

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type DockerComposeYaml struct {
	Version  string                          `json:"version"`
	Services map[string]DockerComposeService `json:"services"`
}

// func (y *DockerComposeYaml) UnmarshalJSON(b []byte) error {
// 	err := json.Unmarshal(b, y)
// 	if err != nil {
// 		return err
// 	}

// 	if y.Version != "3" {
// 		return fmt.Errorf("incompatible version: %s", y.Version)
// 	}

// 	return nil
// }

type DockerComposeServiceCommand []string

func (c *DockerComposeServiceCommand) UnmarshalJSON(b []byte) error {
	var i interface{}
	err := json.Unmarshal(b, &i)
	if err != nil {
		return err
	}

	command := DockerComposeServiceCommand{}
	switch x := i.(type) {
	case []interface{}:
		for _, p := range x {
			command = append(command, p.(string))
		}
		break
	case interface{}:
		re := regexp.MustCompile(`(".+")|('.+')|(\S+)`)
		matches := re.FindAllString(x.(string), -1)
		for _, m := range matches {
			command = append(command, strings.TrimFunc(m, func(r rune) bool {
				return string(r) == `"` || string(r) == `'`
			}))
		}
		break
	default:
		return fmt.Errorf("could not unmarshal command type %T", x)
	}

	*c = command

	return err
}

type DockerComposeServiceEnvironment map[string]string

func (env *DockerComposeServiceEnvironment) UnmarshalJSON(b []byte) error {
	var i interface{}
	err := json.Unmarshal(b, &i)
	if err != nil {
		return err
	}

	environment := DockerComposeServiceEnvironment{}
	switch x := i.(type) {
	case []interface{}:
		for _, e := range x {
			parts := strings.Split(e.(string), "=")
			key := parts[0]
			val := parts[1]
			environment[key] = val
		}
		break
	case map[string]interface{}:
		for k, v := range x {
			environment[k] = v.(string)
		}
		break
	default:
		return fmt.Errorf("could not unmarshal environment type %T", x)
	}

	*env = environment

	return err
}

type DockerComposeServiceNetwork struct {
	Aliases []string `json:"aliases" yaml:"aliases"`
}

type DockerComposeServiceBuild struct {
	Context    string `json:"context" yaml:"context"`
	Dockerfile string `json:"dockerfile" yaml:"dockerfile"`
}

func (c *DockerComposeServiceBuild) UnmarshalJSON(b []byte) error {
	var i interface{}
	err := json.Unmarshal(b, &i)
	if err != nil {
		return err
	}

	switch x := i.(type) {
	case string:
		c.Context = x
		c.Dockerfile = "Dockerfile"
		break
	case map[string]interface{}:
		c.Context = x["context"].(string)
		c.Dockerfile = x["dockerfile"].(string)
		break
	default:
		return fmt.Errorf("could not unmarshal build type %T", x)
	}

	return nil
}

type DockerComposeService struct {
	Image       string                                 `json:"image"`
	Build       DockerComposeServiceBuild              `json:"build"`
	WorkingDir  string                                 `json:"working_dir"`
	Command     DockerComposeServiceCommand            `json:"command"`
	Links       []string                               `json:"links"`
	Environment DockerComposeServiceEnvironment        `json:"environment"`
	Volumes     []string                               `json:"volumes"`
	Expose      []string                               `json:"expose"`
	Networks    map[string]DockerComposeServiceNetwork `json:"networks"`
}

func GetServiceOrder(services map[string]DockerComposeService, serviceOrder []string) []string {
	for serviceName, serviceDef := range services {
		if isIn(serviceName, serviceOrder) {
			break
		}
		for _, linkedService := range serviceDef.Links {
			serviceOrder = getLinkedServiceOrder(linkedService, services, serviceOrder)
		}
		serviceOrder = append(serviceOrder, serviceName)
	}

	for len(services) != len(serviceOrder) {
		serviceOrder = GetServiceOrder(services, serviceOrder)
	}

	return serviceOrder
}

func getLinkedServiceOrder(serviceName string, services map[string]DockerComposeService, serviceOrder []string) []string {
	if isIn(serviceName, serviceOrder) {
		return serviceOrder
	}
	for _, linkedService := range services[serviceName].Links {
		serviceOrder = getLinkedServiceOrder(linkedService, services, serviceOrder)
	}
	return append(serviceOrder, serviceName)
}

func isIn(needle string, haystack []string) bool {
	for _, v := range haystack {
		if needle == v {
			return true
		}
	}
	return false
}
