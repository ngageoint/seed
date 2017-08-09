package dockerHubRegistry

type repositoriesResponse struct {
	Count int
	Next string
	Previous string
	Results []Result
}

type Result struct {
	Name string
}

func (registry *DockerHubRegistry) UserRepositories(user string) ([]string, error) {
	url := registry.url("/v2/repositories/%s/", user)
	repos := make([]string, 0, 10)
	var err error //We create this here, otherwise url will be rescoped with :=
	var response repositoriesResponse
	for err == nil{
		response.Next = ""
		url, err = registry.getDockerHubPaginatedJson(url, &response)
		for _, r := range response.Results {
			repos = append(repos, r.Name)
		}
	}
	if err != ErrNoMorePages {
		return nil, err
	}
	return repos, nil
}
