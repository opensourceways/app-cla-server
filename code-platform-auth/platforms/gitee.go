package platforms

import "github.com/opensourceways/robot-gitee-lib/client"

type giteeClient struct {
	c client.Client
}

func newGiteeClient(accessToken string) *giteeClient {
	cli := client.NewClient(func() []byte { return []byte(accessToken) })

	return &giteeClient{c: cli}
}

func (cli *giteeClient) GetUser() (string, error) {
	v, err := cli.c.GetBot()
	if err != nil {
		return "", err
	}

	return v.Login, err
}

func (cli *giteeClient) ListOrg() ([]string, error) {
	return cli.c.ListOrg()
}
