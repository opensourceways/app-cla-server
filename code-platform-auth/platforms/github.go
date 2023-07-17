package platforms

import "net/http"

const (
	urlToGetGithubUser = "https://api.github.com/user"
	urlToGetGithubOrg  = "https://api.github.com/user/orgs?page=1&per_page=100"
)

func newGithubClient() githubClient {
	return githubClient{}
}

// githubClient
type githubClient struct{}

func (cli githubClient) GetUser(token string) (string, error) {
	req, err := cli.newReq(urlToGetGithubUser, token)
	if err != nil {
		return "", err
	}

	return getUser(req)
}

func (cli githubClient) ListOrg(token string) ([]string, error) {
	req, err := cli.newReq(urlToGetGithubOrg, token)
	if err != nil {
		return nil, err
	}

	return listOrg(req)
}

func (cli githubClient) newReq(url, token string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	return req, nil
}
