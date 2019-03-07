package config

type Root struct {
	Project *ProjectConfig `json:"project"`
	Git     *GitConfig     `json:"git"`

	Parameters []Parameter     `json:"parameters"`
	Plugins    []*PluginConfig `json:"plugins"`
	Stages     []*StageConfig  `json:"stages"`
}

type ProjectConfig struct {
	Logo      *string `json:"logo"`
	TasksPath string  `json:"tasksPath"`
}

type GitConfig struct {
	// Depth     int  `json:"depth"`
	Submodule bool `json:"submodule"`
}

type PluginConfig struct {
	Use       string            `json:"use"`
	Arguments map[string]string `json:"arguments"`
	Events    []string          `json:"events"`
}

type StageConfig struct {
	Name  string   `json:"name"`
	Tasks []string `json:"tasks"`
}
