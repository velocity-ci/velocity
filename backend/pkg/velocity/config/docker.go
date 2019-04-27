package config

type BlueprintDocker struct {
	Registries []BlueprintDockerRegistry `json:"registries"`
}

type BlueprintDockerRegistry struct {
	Address   string            `json:"address"`
	Use       string            `json:"use"`
	Arguments map[string]string `json:"arguments"`
}
