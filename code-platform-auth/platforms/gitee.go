package platforms

import "net/http"

const (
	urlToGetGiteeUser = "https://gitee.com/api/v5/user?access_token="
	urlToGetGiteeOrg  = "https://gitee.com/api/v5/user/orgs?page=1&per_page=100&admin=true&access_token="
)

func newGiteeClient() giteeClient {
	return giteeClient{}
}

// giteeClient
type giteeClient struct{}

func (cli giteeClient) GetUser(token string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, urlToGetGiteeUser+token, nil)
	if err != nil {
		return "", err
	}

	return getUser(req)
}

func (cli giteeClient) ListOrg(token string) ([]string, error) {
	req, err := http.NewRequest(http.MethodGet, urlToGetGiteeOrg+token, nil)
	if err != nil {
		return nil, err
	}

	return listOrg(req)
}
