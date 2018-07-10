package velocity

import (
	"go.uber.org/zap"
)

type RepositoryConfig struct {
	Project ProjectConfig `json:"project" yaml:"project"`
	Git     GitConfig     `json:"git" yaml:"git"`

	Parameters []ParameterConfig `json:"paramaters" yaml:"parameters"`
	Plugins    []PluginConfig    `json:"plugins" yaml:"plugins"`
	Stages     []StageConfig     `json:"stages" yaml:"stages"`
}

type ProjectConfig struct {
	Name      string `json:"name" yaml:"name"`
	Logo      string `json:"logo" yaml:"logo"`
	TasksPath string `json:"tasksPath" yaml:"tasksPath"`
}

type GitConfig struct {
	Depth int `json:"depth" yaml:"depth"`
}

type PluginConfig struct {
	Use       string            `json:"use" yaml:"use"`
	Arguments map[string]string `json:"arguments" yaml:"arguments"`
	Events    []string          `json:"events" yaml:"events"`
}

type StageConfig struct {
	Name  string   `json:"name" yaml:"name"`
	Tasks []string `json:"tasks" yaml:"tasks"`
}

func (t *RepositoryConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var repoConfigMap map[string]interface{}
	err := unmarshal(&repoConfigMap)
	if err != nil {
		GetLogger().Error("unable to marshal repository configuration", zap.Error(err))
		return err
	}

	t.Project = unmarshalProjectYaml(repoConfigMap["project"])
	t.Git = unmarshalGitYaml(repoConfigMap["git"])
	t.Parameters = unmarshalConfigParameters(repoConfigMap["parameters"])
	t.Plugins = unmarshalPluginConfigs(repoConfigMap["plugins"])
	t.Stages = unmarshalStageConfigs(repoConfigMap["stages"])

	return nil
}

func unmarshalProjectYaml(y interface{}) ProjectConfig {
	p := ProjectConfig{
		Name:      "",
		Logo:      "",
		TasksPath: "./tasks",
	}

	switch x := y.(type) {
	case map[interface{}]interface{}:
		if v, ok := x["name"].(string); ok {
			p.Name = v
		}
		if v, ok := x["logo"].(string); ok {
			p.Logo = v
		}
		if v, ok := x["tasksPath"].(string); ok {
			p.TasksPath = v
		}
	}

	return p
}

func unmarshalGitYaml(y interface{}) GitConfig {
	g := GitConfig{
		Depth: 50,
	}
	switch x := y.(type) {
	case map[interface{}]interface{}:
		if v, ok := x["depth"].(int); ok {
			g.Depth = v
		}
	}

	return g
}

func unmarshalPluginConfigs(y interface{}) []PluginConfig {
	pluginConfigs := []PluginConfig{}
	switch x := y.(type) {
	case []interface{}:
		for _, p := range x {
			pluginConfigs = append(pluginConfigs, unmarshalPluginConfig(p))
		}
	}

	return pluginConfigs
}

func unmarshalPluginConfig(y interface{}) PluginConfig {
	pluginConfig := PluginConfig{
		Use:       "",
		Arguments: map[string]string{},
		Events:    []string{},
	}
	switch x := y.(type) {
	case map[interface{}]interface{}:
		if v, ok := x["use"].(string); ok {
			pluginConfig.Use = v
		}
		if v, ok := x["arguments"].(map[interface{}]interface{}); ok {
			for k, v := range v {
				sK, okk := k.(string)
				sV, okv := v.(string)
				if okk && okv {
					pluginConfig.Arguments[sK] = sV
				}
			}
		}
		if v, ok := x["events"].([]interface{}); ok {
			for _, v := range v {
				if v, ok := v.(string); ok {
					pluginConfig.Events = append(pluginConfig.Events, v)
				}
			}
		}
	}

	return pluginConfig
}

func unmarshalStageConfigs(y interface{}) []StageConfig {
	stageConfigs := []StageConfig{}

	switch x := y.(type) {
	case []interface{}:
		for _, s := range x {
			stageConfigs = append(stageConfigs, unmarshalStageConfig(s))
		}
	}

	return stageConfigs
}

func unmarshalStageConfig(y interface{}) StageConfig {
	stageConfig := StageConfig{
		Name:  "",
		Tasks: []string{},
	}
	switch x := y.(type) {
	case map[interface{}]interface{}:
		if v, ok := x["name"].(string); ok {
			stageConfig.Name = v
		}
		if v, ok := x["tasks"].([]interface{}); ok {
			for _, v := range v {
				if v, ok := v.(string); ok {
					stageConfig.Tasks = append(stageConfig.Tasks, v)
				}
			}
		}
	}
	return stageConfig
}
