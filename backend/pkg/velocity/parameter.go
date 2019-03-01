package velocity

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/gosimple/slug"
)

type Parameter struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	IsSecret bool   `json:"isSecret"`
}

type ParameterConfig interface {
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

func NewBasicParameter() *BasicParameter {
	return &BasicParameter{
		Type: "basic",
	}
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

type DerivedParameter struct {
	Type      string            `json:"type"`
	Use       string            `json:"use" yaml:"use"`
	Secret    bool              `json:"secret" yaml:"secret"`
	Arguments map[string]string `json:"arguments" yaml:"arguments"`
	Exports   map[string]string `json:"exports" yaml:"exports"`
	// Timeout   uint64
}

func NewDerivedParameter() *DerivedParameter {
	return &DerivedParameter{
		Type: "derived",
	}
}

func (p DerivedParameter) GetInfo() string {
	return p.Use
}

func getBinary(projectRoot, u string, writer io.Writer) (binaryLocation string, _ error) {

	parsedURL, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	binaryLocation = fmt.Sprintf("%s/.velocityci/plugins/%s", projectRoot, slug.Make(parsedURL.Path))

	if _, err := os.Stat(binaryLocation); os.IsNotExist(err) {
		GetLogger().Debug("downloading binary", zap.String("from", u), zap.String("to", binaryLocation))
		writer.Write([]byte(fmt.Sprintf("Downloading binary: %s", parsedURL.String())))
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
		writer.Write([]byte(fmt.Sprintf(
			"Downloaded binary: %s to %s. %d bytes",
			parsedURL.String(),
			binaryLocation,
			size,
		)))

		GetLogger().Debug("downloaded binary", zap.String("from", u), zap.String("to", binaryLocation), zap.Int64("bytes", size))
		outFile.Chmod(os.ModePerm)
	}

	return binaryLocation, nil
}

func (p DerivedParameter) GetParameters(writer io.Writer, t *Task, backupResolver BackupResolver) (r []Parameter, _ error) {

	// Download binary from use:
	bin, err := getBinary(t.ProjectRoot, p.Use, writer)
	if err != nil {
		return r, err
	}
	cmd := []string{bin}

	// Process arguments
	for k, v := range p.Arguments {
		cmd = append(cmd, fmt.Sprintf("-%s=%s", k, v))
	}

	// Run binary
	s := runCmd(BlankWriter{}, cmd, os.Environ())
	if s.Error != nil {
		return r, s.Error
	}
	var dOutput derivedOutput
	json.Unmarshal([]byte(s.Stdout[0]), &dOutput)

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
				Name:     getExportedParameterName(p.Exports, paramName),
				Value:    val,
				IsSecret: dOutput.Secret,
			})
		}
	} else {
		return r, fmt.Errorf("binary %s: %s", dOutput.State, dOutput.Error)
	}

	return r, nil
}

func getExportedParameterName(pMapping map[string]string, exportedParam string) string {
	if val, ok := pMapping[exportedParam]; ok {
		return val
	}

	return exportedParam
}

type derivedOutput struct {
	Secret  bool              `json:"secret"`
	Exports map[string]string `json:"exports"`
	Expires time.Time         `json:"expires"`
	Error   string            `json:"error"`
	State   string            `json:"state"`
}

func unmarshalConfigParameter(b []byte) (p ParameterConfig, err error) {
	var m map[string]interface{}
	err = json.Unmarshal(b, &m)
	if err != nil {
		return p, err
	}

	if _, ok := m["use"]; ok { // derived
		p = NewDerivedParameter()
	} else if _, ok := m["name"]; ok { // basic
		p = NewBasicParameter()
	}

	err = json.Unmarshal(b, p)

	return p, err
}
