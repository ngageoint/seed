package registry

type repositoriesResponse struct {
	Count int
	Next string
	Previous string
	Results []Result
	Repositories []string `json:"repositories"`
}

type Result struct {
	Name string
}

func (registry *Registry) Repositories() ([]string, error) {
	url := registry.url("/v2/_catalog")
	repos := make([]string, 0, 10)
	var err error //We create this here, otherwise url will be rescoped with :=
	var response repositoriesResponse
	for {
		registry.Logf("registry.repositories url=%s", url)
		url, err = registry.getPaginatedJson(url, &response)
		switch err {
		case ErrNoMorePages:
			repos = append(repos, response.Repositories...)
			return repos, nil
		case nil:
			repos = append(repos, response.Repositories...)
			continue
		default:
			return nil, err
		}
	}
}

func (registry *Registry) UserRepositories(user string) ([]string, error) {
	url := registry.url("/v2/repositories/%s/", user)
	repos := make([]string, 0, 10)
	var err error //We create this here, otherwise url will be rescoped with :=
	var response repositoriesResponse
	for err == nil{
		//registry.Logf("registry.repositories url=%s", url)
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
