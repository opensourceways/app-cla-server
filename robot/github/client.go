package github

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	sdk "github.com/google/go-github/v36/github"
	"golang.org/x/oauth2"

	"github.com/opensourceways/app-cla-server/robot/cla"
)

func newGithubClient(getToken func() []byte) *sdk.Client {
	acToken := string(getToken())
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: acToken})
	tc := oauth2.NewClient(context.Background(), ts)
	return sdk.NewClient(tc)
}

type client struct {
	c        *sdk.Client
	isLitePR func(email string, name string) bool
}

func (cl client) AddPRLabel(pr cla.PRInfo, label string) error {
	_, _, err := cl.c.Issues.AddLabelsToIssue(context.Background(), pr.Org, pr.Repo, pr.Number, []string{label})
	return err
}

func (cl client) RemovePRLabel(pr cla.PRInfo, label string) error {
	r, err := cl.c.Issues.RemoveLabelForIssue(context.Background(), pr.Org, pr.Repo, pr.Number, label)
	if err != nil && r != nil && r.StatusCode == 404 {
		return nil
	}
	return err
}

func (cl client) CreatePRComment(pr cla.PRInfo, comment string) error {
	ic := sdk.IssueComment{
		Body: sdk.String(comment),
	}
	_, _, err := cl.c.Issues.CreateComment(context.Background(), pr.Org, pr.Repo, pr.Number, &ic)
	return err
}

func (cl client) DeletePRComment(pr cla.PRInfo, needDelete func(string) bool) error {
	for _, item := range cl.listPRComment(pr) {
		if needDelete(item.GetBody()) {
			cl.c.Issues.DeleteComment(context.Background(), pr.Org, pr.Repo, item.GetID())
		}
	}

	return nil
}

func (cl client) listPRComment(pr cla.PRInfo) []*sdk.IssueComment {
	comments := []*sdk.IssueComment{}

	opt := &sdk.IssueListCommentsOptions{}
	opt.Page = 1

	for {
		v, resp, err := cl.c.Issues.ListComments(context.Background(), pr.Org, pr.Repo, pr.Number, opt)
		if err != nil {
			break
		}

		comments = append(comments, v...)

		link := parseLinks(resp.Header.Get("Link"))["next"]
		if link == "" {
			break
		}

		pagePath, err := url.Parse(link)
		if err != nil {
			// return fmt.Errorf("failed to parse 'next' link: %v", err)
			break
		}

		p := pagePath.Query().Get("page")
		if p == "" {
			// return fmt.Errorf("failed to get 'page' on link: %s", p)
			break
		}

		page, err := strconv.Atoi(p)
		if err != nil {
			// return err
			break
		}
		opt.Page = page
	}

	return comments
}

func (cl client) GetUnsignedCommits(pr cla.PRInfo, commiterAsAuthor bool, isSigned func(string) bool) (map[string]string, error) {
	commits, err := cl.listPRCommits(pr)
	if err != nil {
		return nil, err
	}

	if len(commits) == 0 {
		return nil, fmt.Errorf("empty pr")
	}

	authorEmailOfCommit := func(c *sdk.RepositoryCommit) string {
		return cl.getAuthorOfCommit(c, commiterAsAuthor)
	}

	unsigned := make(map[string]string)
	signed := 0
	update := func(c *sdk.RepositoryCommit, b bool) {
		if b {
			signed++
		} else {
			unsigned[c.GetSHA()] = c.GetCommit().GetMessage()
		}
	}

	result := make(map[string]bool)

	for _, c := range commits {
		email := authorEmailOfCommit(c)
		if email == "" {
			update(c, false)
			continue
		}

		email = strings.Trim(email, " \"")

		if v, ok := result[email]; ok {
			update(c, v)
		} else {
			b := isSigned(email)
			result[email] = b
			update(c, b)
		}
	}

	if _, ok := unsigned[""]; ok {
		return nil, fmt.Errorf("invalid commit exists")
	}

	if len(commits) != signed+len(unsigned) {
		// it is impossible except that there are two or more commits has same SHA.
		return nil, fmt.Errorf("impossible")
	}

	return unsigned, nil
}

func (cl client) listPRCommits(pr cla.PRInfo) ([]*sdk.RepositoryCommit, error) {
	commits := []*sdk.RepositoryCommit{}

	f := func() error {
		opt := &sdk.ListOptions{}
		opt.Page = 1

		for {
			v, resp, err := cl.c.PullRequests.ListCommits(context.Background(), pr.Org, pr.Repo, pr.Number, nil)
			if err != nil {
				return err
			}

			commits = append(commits, v...)

			link := parseLinks(resp.Header.Get("Link"))["next"]
			if link == "" {
				break
			}

			pagePath, err := url.Parse(link)
			if err != nil {
				return fmt.Errorf("failed to parse 'next' link: %v", err)
			}

			p := pagePath.Query().Get("page")
			if p == "" {
				return fmt.Errorf("failed to get 'page' on link: %s", p)
			}

			page, err := strconv.Atoi(p)
			if err != nil {
				return err
			}

			opt.Page = page
		}

		return nil
	}

	err := f()
	return commits, err
}

func (cl client) getAuthorOfCommit(c *sdk.RepositoryCommit, byCommitter bool) string {
	commit := c.GetCommit()

	if byCommitter {
		committer := commit.GetCommitter()
		if !cl.isLitePR(committer.GetEmail(), committer.GetName()) {
			return committer.GetEmail()
		}
	}

	return commit.GetAuthor().GetEmail()
}
