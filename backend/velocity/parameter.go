package velocity

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/docker/go/canonical/json"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
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

func (p DerivedParameter) GetParameters(writer io.Writer, t *Task, backupResolver BackupResolver) ([]Parameter, error) {
	env := []string{}
	for k, v := range p.Arguments {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	containerConfig := &container.Config{
		Image: p.Use,
		Env:   env,
	}

	var wg sync.WaitGroup
	cli, _ := client.NewEnvClient()
	ctx := context.Background()

	networkResp, err := cli.NetworkCreate(ctx, t.RunID, types.NetworkCreate{
		Labels: map[string]string{"owner": "velocity-ci"},
	})
	if err != nil {
		log.Println(err)
	}
	blankEmitter := NewBlankEmitter()
	w := blankEmitter.GetStreamWriter("setup")
	sR := newServiceRunner(
		cli,
		ctx,
		w,
		&wg,
		map[string]Parameter{},
		fmt.Sprintf("%s-%s", t.RunID, "dParam"),
		p.Use,
		nil,
		containerConfig,
		nil,
		nil,
		networkResp.ID,
	)

	sR.PullOrBuild(t.Docker.Registries)
	sR.Create()
	stopServicesChannel := make(chan string, 32)
	wg.Add(1)
	go sR.Run(stopServicesChannel)
	_ = <-stopServicesChannel

	logsResp, err := sR.dockerCli.ContainerLogs(
		sR.context,
		sR.containerID,
		types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: false},
	)
	if err != nil {
		log.Printf("param container %s logs err: %s", sR.containerID, err)
	}
	headerBytes := make([]byte, 8)
	logsResp.Read(headerBytes)
	content, _ := ioutil.ReadAll(logsResp)
	logsResp.Close()
	sR.Stop()
	wg.Wait()
	err = cli.NetworkRemove(ctx, networkResp.ID)
	if err != nil {
		log.Printf("network %s remove err: %s", networkResp.ID, err)
	}

	params := []Parameter{}
	var dOutput derivedOutput
	json.Unmarshal(content, &dOutput)
	if dOutput.State == "warning" {
		for paramName := range dOutput.Exports {
			val, err := backupResolver.Resolve(paramName)
			if err != nil {
				return []Parameter{}, err
			}
			params = append(params, Parameter{
				Name:     paramName,
				Value:    val,
				IsSecret: dOutput.Secret,
			})
		}
	} else if dOutput.State == "success" {
		for paramName, val := range dOutput.Exports {
			params = append(params, Parameter{
				Name:     paramName,
				Value:    val,
				IsSecret: dOutput.Secret,
			})
		}
	} else { // failed
		return []Parameter{}, errors.New(dOutput.Error)
	}

	return params, nil
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
