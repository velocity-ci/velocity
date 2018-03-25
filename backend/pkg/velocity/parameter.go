package velocity

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gosimple/slug"
)

type Parameter struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	IsSecret bool   `json:"isSecret"`
}

type ConfigParameter interface {
	GetInfo() string
	GetParameters(writer io.Writer, t *Task, backupResolver BackupResolver) ([]Parameter, error)
}

type BackupResolver interface {
	Resolve(paramName string) (string, error)
}

type BasicParameter struct {
	Type         string   `json:"type"`
	Name         string   `json:"name" yaml:"name"`
	Default      string   `json:"default" yaml:"default"`
	OtherOptions []string `json:"otherOptions" yaml:"otherOptions"`
	Secret       bool     `json:"secret" yaml:"secret"`
	Value        string   `json:"value"`
}

func (p BasicParameter) GetInfo() string {
	return p.Name
}

func (p BasicParameter) GetParameters(writer io.Writer, t *Task, backupResolver BackupResolver) ([]Parameter, error) {
	v := p.Default
	if len(p.Value) > 0 {
		v = p.Value
	} else {
		val, err := backupResolver.Resolve(p.Name)
		if err != nil {
			return []Parameter{}, err
		}
		v = val
	}
	return []Parameter{
		{
			Name:     p.Name,
			Value:    v,
			IsSecret: p.Secret,
		},
	}, nil
}

func (p *BasicParameter) UnmarshalYamlInterface(y map[interface{}]interface{}) error {
	p.Type = "basic"
	switch x := y["name"].(type) {
	case interface{}:
		p.Name = x.(string)
		break
	}
	switch x := y["default"].(type) {
	case interface{}:
		p.Default = x.(string)
		break
	}
	switch x := y["secret"].(type) {
	case interface{}:
		p.Secret = x.(bool)
		break
	}

	p.OtherOptions = []string{}
	switch x := y["otherOptions"].(type) {
	case []interface{}:
		for _, o := range x {
			p.OtherOptions = append(p.OtherOptions, o.(string))
		}
		break
	}

	return nil
}

type DerivedParameter struct {
	Type      string            `json:"type"`
	Use       string            `json:"use" yaml:"use"`
	Secret    bool              `json:"secret" yaml:"secret"`
	Arguments map[string]string `json:"arguments" yaml:"arguments"`
	Exports   map[string]string `json:"exports" yaml:"exports"`
	// Timeout   uint64
}

func (p DerivedParameter) GetInfo() string {
	return p.Use
}

func getBinary(u string) (binaryLocation string, _ error) {

	parsedURL, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	binaryLocation = fmt.Sprintf("%s/.velocityci/plugins/%s", wd, slug.Make(parsedURL.Path))

	if _, err := os.Stat(binaryLocation); os.IsNotExist(err) {
		logrus.Infof("downloading %s to %s", u, binaryLocation)
		outFile, err := os.Create(binaryLocation)
		if err != nil {
			return "", err
		}
		defer outFile.Close()
		resp, err := http.Get(u)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		size, err := io.Copy(outFile, resp.Body)
		if err != nil {
			return "", err
		}
		logrus.Infof("downloaded %d bytes for %s to %s", size, u, binaryLocation)
		outFile.Chmod(os.ModePerm)
	}

	return binaryLocation, nil
}

func (p DerivedParameter) GetParameters(writer io.Writer, t *Task, backupResolver BackupResolver) (r []Parameter, _ error) {

	// Download binary from use:
	bin, err := getBinary(p.Use)
	if err != nil {
		return r, err
	}

	// Process arguments
	args := []string{}
	for k, v := range p.Arguments {
		args = append(args, fmt.Sprintf("-%s=%s", k, v))
	}

	cmd := exec.Command(bin, args...)
	cmd.Env = os.Environ()

	// Run binary
	cmdOutBytes, err := cmd.Output()
	if err != nil {
		return r, err
	}
	var dOutput derivedOutput
	json.Unmarshal(cmdOutBytes, &dOutput)

	if dOutput.State == "warning" {
		for paramName := range dOutput.Exports {
			val, err := backupResolver.Resolve(paramName)
			if err != nil {
				return r, err
			}
			r = append(r, Parameter{
				Name:     paramName,
				Value:    val,
				IsSecret: dOutput.Secret,
			})
		}
	} else if dOutput.State == "success" {
		for paramName, val := range dOutput.Exports {
			r = append(r, Parameter{
				Name:     paramName,
				Value:    val,
				IsSecret: dOutput.Secret,
			})
		}
	} else {
		return r, fmt.Errorf("binary %s: %s", dOutput.State, dOutput.Error)
	}

	return r, nil
}

func (p *DerivedParameter) UnmarshalYamlInterface(y map[interface{}]interface{}) error {

	p.Type = "derived"

	switch x := y["use"].(type) {
	case interface{}:
		p.Use = x.(string)
		break
	}
	switch x := y["secret"].(type) {
	case interface{}:
		p.Secret = x.(bool)
		break
	}
	p.Arguments = map[string]string{}
	switch x := y["arguments"].(type) {
	case map[interface{}]interface{}:
		for k, v := range x {
			p.Arguments[k.(string)] = v.(string)
		}
		break
	}
	p.Exports = map[string]string{}
	switch x := y["exports"].(type) {
	case map[interface{}]interface{}:
		for k, v := range x {
			p.Exports[k.(string)] = v.(string)
		}
		break
	}

	return nil
}

type derivedOutput struct {
	Secret  bool              `json:"secret"`
	Exports map[string]string `json:"exports"`
	Expires time.Time         `json:"expires"`
	Error   string            `json:"error"`
	State   string            `json:"state"`
}
