package config

type TaskDocker struct {
	Registries []TaskDockerRegistry `json:"registries"`
}

type TaskDockerRegistry struct {
	Address   string            `json:"address"`
	Use       string            `json:"use"`
	Arguments map[string]string `json:"arguments"`
}
