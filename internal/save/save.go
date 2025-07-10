package save

type SaveOptions struct {
	OutputDir         string `json:"output_dir"`
	DockerComposePath string `json:"docker_compose_path"`
	WorkDir           string `json:"work_dir"`
}

type SaveInterface interface {
	LoadComposeFile() error
	SaveProject() error
	SaveComposeFile() error
	LoadDockerProject() error
}

type SaveClient struct {
	Options SaveOptions
	SaveInterface
}

func NewSaveClient(options SaveOptions) *SaveClient {
	return &SaveClient{
		Options: options,
	}
}

func (c *SaveClient) LoadComposeFile() error {
	// Implement logic to load the Docker Compose file
	return nil
}

func (c *SaveClient) SaveProject() error {
	// Implement logic to save the Docker Compose project
	return nil
}
func (c *SaveClient) SaveComposeFile() error {
	// Implement logic to save the Docker Compose file
	return nil
}
func (c *SaveClient) LoadDockerProject() error {
	// Implement logic to load the Docker project
	return nil
}
