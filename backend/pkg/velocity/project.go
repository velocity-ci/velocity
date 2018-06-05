package velocity

type ProjectConfig struct {
	Logo      string `json:"logo"`
	TasksPath string `json:"tasksPath"`

	Git GitConfig `json:"git"`

	Parameters []ConfigParameter `json:"paramaters"`
	Plugins    []PluginConfig    `json:"plugins"`
	Stages     []StageConfig     `json:"stages"`
}

type GitConfig struct {
	Depth int `json:"depth"`
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
