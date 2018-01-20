package velocity

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"sync"
	"time"

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
	GetParameters(writer io.Writer, runID string) ([]Parameter, error)
}

type BasicParameter struct {
	Name         string   `json:"name" yaml:"name"`
	Default      string   `json:"default" yaml:"default"`
	OtherOptions []string `json:"otherOptions" yaml:"otherOptions"`
	Secret       bool     `json:"secret" yaml:"secret"`
	Value        string   `json:"value"`
}

func (p BasicParameter) GetParameters(writer io.Writer, runID string) ([]Parameter, error) {
	v := p.Default
	if len(p.Value) > 0 {
		v = p.Value
	}
	return []Parameter{
		Parameter{
			Name:     p.Name,
			Value:    v,
			IsSecret: p.Secret,
		},
	}, nil
}

type DerivedParameter struct {
	Use       string            `json:"use" yaml:"use"`
	Secret    bool              `json:"secret" yaml:"secret"`
	Arguments map[string]string `json:"arguments" yaml:"arguments"`
	Exports   map[string]string `json:"exports" yaml:"exports"`
	// Timeout   uint64
}

func (p DerivedParameter) GetParameters(writer io.Writer, runID string) ([]Parameter, error) {
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

	networkResp, err := cli.NetworkCreate(ctx, runID, types.NetworkCreate{
		Labels: map[string]string{"owner": "velocity-ci"},
	})
	if err != nil {
		log.Println(err)
	}

	sR := newServiceRunner(
		cli,
		ctx,
		writer,
		&wg,
		map[string]Parameter{},
		fmt.Sprintf("%s-%s", runID, "dParam"),
		p.Use,
		nil,
		containerConfig,
		nil,
		nil,
		networkResp.ID,
	)

	sR.PullOrBuild()
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
	content, _ := ioutil.ReadAll(logsResp)
	logsResp.Close()

	fmt.Println(content)

	sR.Stop()
	wg.Wait()
	err = cli.NetworkRemove(ctx, networkResp.ID)
	if err != nil {
		log.Printf("network %s remove err: %s", networkResp.ID, err)
	}

	// exitCode := sR.exitCode

	return []Parameter{}, nil
}

type derivedOutput struct {
	Secret  bool              `json:"secret"`
	Exports map[string]string `json:"exports"`
	Expires time.Time         `json:"expires"`
	Error   string            `json:"error"`
	State   string            `json:"state"`
}
