package config

type blueprintDocker struct {
	Registries []blueprintDockerRegistry `json:"registries"`
}

type blueprintDockerRegistry struct {
	Address   string            `json:"address"`
	Use       string            `json:"use"`
	Arguments map[string]string `json:"arguments"`
}
