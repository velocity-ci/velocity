package velocity

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

type Plugin struct {
	BaseStep
	Use       string            `json:"use" yaml:"use"`
	Arguments map[string]string `json:"arguments" yaml:"arguments"`
}

func (p Plugin) GetDetails() string {
	return fmt.Sprintf("use: %s", p.Use)
}

func (p *Plugin) Execute(emitter Emitter, t *Task) error {

	type output struct {
		State string `json:"state"`
		Error string `json:"error"`
	}

	writer := emitter.GetStreamWriter("plugin")
	defer writer.Close()
	writer.SetStatus(StateRunning)

	bin, err := getBinary(p.Use)
	if err != nil {
		return err
	}

	env := []string{}
	for k, v := range p.Arguments {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(), env...)

	cmdOutBytes, err := cmd.Output()
	if err != nil {
		return err
	}
	var dOutput output
	json.Unmarshal(cmdOutBytes, &dOutput)

	if dOutput.State != "success" {
		writer.SetStatus("failed")
		writer.Write([]byte(fmt.Sprintf("\n%s\n### FAILED (error: %s)\x1b[0m", errorANSI, dOutput.Error)))
		return fmt.Errorf("error: %s", dOutput.Error)
	}

	writer.SetStatus("success")
	writer.Write([]byte(fmt.Sprintf("\n%s\n### SUCCESS \x1b[0m", successANSI)))
	return nil
}

func (p Plugin) Validate(params map[string]Parameter) error {

	return nil
}

func (p *Plugin) SetParams(params map[string]Parameter) error {

	return nil
}
