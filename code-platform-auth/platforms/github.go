package platforms

import "github.com/opensourceways/robot-github-lib/client"

type githubClient struct {
	c client.Client
}

func newGithubClient(accessToken string) *githubClient {
	cli := client.NewClient(func() []byte { return []byte(accessToken) })

	return &githubClient{c: cli}
}

func (cli *githubClient) GetUser() (string, error) {
	return cli.c.GetBot()
}

func (cli *githubClient) ListOrg() ([]string, error) {
	return cli.c.ListOrg()
}
