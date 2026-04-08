package discover

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type DockerProvider struct {
	verbose bool
}

func NewDockerProvider(verbose bool) *DockerProvider {
	return &DockerProvider{verbose: verbose}
}

func (p *DockerProvider) Name() string {
	return "docker"
}

func (p *DockerProvider) IsAvailable() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

type dockerContainer struct {
	ID    string `json:"Id"`
	Name  string `json:"Name"`
	State struct {
		Pid    int    `json:"Pid"`
		Status string `json:"Status"`
	} `json:"State"`
	Config struct {
		Image  string `json:"Image"`
		Labels map[string]string `json:"Labels"`
	} `json:"Config"`
	Image string `json:"Image"`
}

func (p *DockerProvider) Discover() ([]*Container, error) {
	// List all running container IDs
	out, err := exec.Command("docker", "ps", "-q", "--no-trunc").Output()
	if err != nil {
		return nil, fmt.Errorf("docker ps failed: %v", err)
	}

	ids := strings.Fields(strings.TrimSpace(string(out)))
	if len(ids) == 0 {
		return nil, nil
	}

	// Inspect all at once
	args := append([]string{"inspect"}, ids...)
	out, err = exec.Command("docker", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("docker inspect failed: %v", err)
	}

	var containers []dockerContainer
	if err := json.Unmarshal(out, &containers); err != nil {
		return nil, fmt.Errorf("failed to parse docker inspect output: %v", err)
	}

	var results []*Container
	for _, c := range containers {
		if c.State.Status != "running" || c.State.Pid <= 0 {
			continue
		}

		name := strings.TrimPrefix(c.Name, "/")

		results = append(results, &Container{
			PID:           c.State.Pid,
			Runtime:       "docker",
			MountNS:       getMountNamespace(c.State.Pid),
			Cgroup:        getCgroupInfo(c.State.Pid),
			ContainerID:   c.ID[:12],
			ContainerName: name,
			ImageRef:      c.Config.Image,
			ImageID:       c.Image,
		})
	}

	return results, nil
}
