package domain

type Config struct {
	ConfigVersion int    `json:"config_version"`
	DockerHost    string `json:"docker_host"`
	Theme         string `json:"theme"`
}
