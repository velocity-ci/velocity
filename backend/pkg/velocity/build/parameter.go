package build

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
	"github.com/velocity-ci/velocity/backend/pkg/exec"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/config"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/logging"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/out"
)

func getSecrets(params map[string]*Parameter) (r []string) {
	for _, p := range params {
		if p.IsSecret {
			r = append(r, p.Value)
		}
	}

	return r
}

func resolveConfigParameter(
	p config.Parameter,
	bR BackupResolver,
	projectRoot string,
	writer io.Writer,
) (parameters []*Parameter, err error) {
	// resolve parameter value at build time
	switch x := p.(type) {
	case config.ParameterBasic:
		writer.Write([]byte(fmt.Sprintf("-> resolving parameter %s", x.Name)))
		return resolveConfigParameterBasic(x, bR)
	case config.ParameterDerived:
		writer.Write([]byte(fmt.Sprintf("-> resolving parameter %s", x.Use)))
		return resolveConfigParameterDerived(x, bR, projectRoot, writer)
	}

	return parameters, err
}

func resolveConfigParameterBasic(p config.ParameterBasic, backupResolver BackupResolver) (parameters []*Parameter, err error) {
	v := p.Default
	val, err := backupResolver.Resolve(p.Name)
	if err != nil {
		return nil, err
	}
	v = val
	return []*Parameter{&Parameter{
		Name:     p.Name,
		Value:    v,
		IsSecret: p.Secret,
	}}, err
}

func resolveConfigParameterDerived(
	p config.ParameterDerived,
	backupResolver BackupResolver,
	projectRoot string,
	writer io.Writer,
) (parameters []*Parameter, err error) {
	// Download binary from use:
	bin, err := getBinary(projectRoot, p.Use, writer)
	if err != nil {
		return parameters, err
	}
	cmd := []string{bin}

	// Process arguments
	for k, v := range p.Arguments {
		cmd = append(cmd, fmt.Sprintf("-%s=%s", k, v))
	}

	// Run binary
	s := exec.Run(cmd, "", os.Environ(), out.BlankWriter{})
	if s.Error != nil {
		return parameters, s.Error
	}
	var dOutput derivedOutput
	json.Unmarshal([]byte(s.Stdout[0]), &dOutput)

	if dOutput.State == "warning" {
		for paramName := range dOutput.Exports {
			val, err := backupResolver.Resolve(paramName)
			if err != nil {
				return parameters, err
			}
			parameters = append(parameters, &Parameter{
				Name:     paramName,
				Value:    val,
				IsSecret: dOutput.Secret,
			})
		}
	} else if dOutput.State == "success" {
		for paramName, val := range dOutput.Exports {
			parameters = append(parameters, &Parameter{
				Name:     getExportedParameterName(p.Exports, paramName),
				Value:    val,
				IsSecret: dOutput.Secret,
			})
		}
	} else {
		return parameters, fmt.Errorf("binary %s: %s", dOutput.State, dOutput.Error)
	}

	return parameters, nil
}

type Parameter struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	IsSecret bool   `json:"isSecret"`
}

type BackupResolver interface {
	Resolve(paramName string) (string, error)
}

func getBinary(projectRoot, u string, writer io.Writer) (binaryLocation string, _ error) {

	parsedURL, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	binaryLocation = fmt.Sprintf("%s/.velocityci/plugins/%s", projectRoot, slug.Make(parsedURL.Path))

	if _, err := os.Stat(binaryLocation); os.IsNotExist(err) {
		logging.GetLogger().Debug("downloading binary", zap.String("from", u), zap.String("to", binaryLocation))
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

		logging.GetLogger().Debug("downloaded binary", zap.String("from", u), zap.String("to", binaryLocation), zap.Int64("bytes", size))
		outFile.Chmod(os.ModePerm)
	}

	return binaryLocation, nil
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