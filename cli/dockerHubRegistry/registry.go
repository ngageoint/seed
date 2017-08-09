package dockerHubRegistry

import (
	"fmt"
	"net/http"
	"strings"
)

type DockerHubRegistry struct {
	URL    string
	Client *http.Client
}

func New(registryUrl string) (*DockerHubRegistry, error) {
	url := strings.TrimSuffix(registryUrl, "/")
	registry := &DockerHubRegistry{
		URL: url,
		Client: &http.Client{},
	}

	return registry, nil
}

func (r *DockerHubRegistry) url(pathTemplate string, args ...interface{}) string {
	pathSuffix := fmt.Sprintf(pathTemplate, args...)
	url := fmt.Sprintf("%s%s", r.URL, pathSuffix)
	return url
}
