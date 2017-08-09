package dockerHubRegistry

import (
	"encoding/json"
	"errors"
	"net/http"
)

var (
	ErrNoMorePages = errors.New("No more pages")
)

// getDockerHubPaginatedJson works with the list of repositories for a user
// returned by docker hub. accepts a string and a pointer, and returns the
// next page URL while updating pointed-to variable with a parsed JSON
// value. When there are no more pages it returns `ErrNoMorePages`.
func (registry *DockerHubRegistry) getDockerHubPaginatedJson(url string, response interface{}) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(response)
	r := response.(*repositoriesResponse)
	if err != nil {
		return "", err
	}
	if r.Next == "" {
		err = ErrNoMorePages
	}
	return r.Next, err
}
