package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sync"

	"github.com/astaxie/beego"

	sdk "github.com/google/go-github/v36/github"
	"github.com/opensourceways/robot-gitee-plugin-lib/config"
	"github.com/opensourceways/robot-gitee-plugin-lib/secret"
	"k8s.io/apimachinery/pkg/util/sets"

	robotConfig "github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/robot/cla"
)

func ValidateWebhook(payload []byte, getHeader func(string) string) (string, string, int, error) {
	if robot == nil {
		return "", "", 400, fmt.Errorf("unsupported")
	}

	return robot.access.checkWebhook(payload, getHeader)
}

func Handle(eventType string, payload []byte) error {
	if robot == nil {
		return errors.New("unsupported")
	}

	return robot.dispatch(eventType, payload)
}

func InitGithubRobot(endpoint string, cfgs []robotConfig.PlatformRobotConfig) error {
	i, ok := robotConfig.FindPlatformRobotConfig("github", cfgs)
	if !ok {
		return nil
	}

	cfg := cfgs[i]

	agent := config.NewConfigAgent(func() config.PluginConfig {
		return new(cla.Configuration)
	})
	if err := agent.Start(cfg.CLARepoConfigFile); err != nil {
		return err
	}

	secretAgent := new(secret.Agent)
	err := secretAgent.Start([]string{
		cfg.HMacSecretFile,
		cfg.RobotTokenFile,
	})
	if err != nil {
		return err
	}

	cl := client{
		c:        newGithubClient(secretAgent.GetTokenGenerator(cfg.RobotTokenFile)),
		isLitePR: cfg.LitePRCommitter.IsLitePR,
	}

	signURL, err := url.Parse(endpoint)
	if err != nil {
		return err
	}

	robot = &server{
		handler: cla.NewCLA(
			func() *cla.Configuration {
				_, v := agent.GetConfig()
				return v.(*cla.Configuration)
			},
			cl, signURL, cfg.FAQOfCheckingByAuthor, cfg.FAQOfCheckingByCommitter,
		),
		secretAgent:   secretAgent,
		claRepoConfig: &agent,
		access:        access{getHamc: secretAgent.GetTokenGenerator(cfg.HMacSecretFile)},
	}
	return nil
}

var robot *server

type server struct {
	claRepoConfig *config.ConfigAgent
	secretAgent   *secret.Agent

	wg      sync.WaitGroup
	handler cla.Handler
	access  access
}

func (r *server) dispatch(eventType string, payload []byte) error {
	switch eventType {
	case "issue_comment":
		var ic sdk.IssueCommentEvent
		if err := json.Unmarshal(payload, &ic); err != nil {
			return err
		}

		if ic.GetAction() == "created" && cla.CheckCLARe.MatchString(ic.GetComment().GetBody()) {
			r.wg.Add(1)
			go r.handleIssueCommentEvent(&ic)
		}

	case "pull_request":
		var pr sdk.PullRequestEvent
		if err := json.Unmarshal(payload, &pr); err != nil {
			return err
		}
		a := pr.GetAction()
		if a == "synchronize" || a == "opened" {
			r.wg.Add(1)
			go r.handlePullRequestEvent(&pr)
		}
	}

	return nil
}

func (s *server) handleIssueCommentEvent(e *sdk.IssueCommentEvent) {
	defer s.wg.Done()

	pr := cla.PRInfo{
		Org:    e.GetRepo().GetOwner().GetLogin(),
		Repo:   e.GetRepo().GetName(),
		Number: e.GetIssue().GetNumber(),
	}

	labels := sets.NewString()
	for _, item := range e.GetIssue().Labels {
		labels.Insert(item.GetName())
	}

	signed, err := s.handler.Handle(pr, labels)
	if err != nil {
		// TODO: how to deal with error
		beego.Error(
			fmt.Sprintf(
				"error to handle issue comment for %s, err:%s",
				pr.String(), err.Error(),
			),
		)
	}

	s.handStatus(signed)
}

func (s *server) handlePullRequestEvent(e *sdk.PullRequestEvent) {
	defer s.wg.Done()

	pr := cla.PRInfo{
		Org:    e.GetRepo().GetOwner().GetLogin(),
		Repo:   e.GetRepo().GetName(),
		Number: e.GetNumber(),
	}

	labels := sets.NewString()
	for _, item := range e.GetPullRequest().Labels {
		labels.Insert(item.GetName())
	}

	signed, err := s.handler.Handle(pr, labels)
	if err != nil {
		// TODO: how to deal with error
		beego.Error(
			fmt.Sprintf(
				"error to handle pr event for %s, err:%s",
				pr.String(), err.Error(),
			),
		)
	}

	s.handStatus(signed)
}

func (s *server) handStatus(signed bool) {
	// TODO: Create Or Update status. Maybe need sign url to be shown in the status.
}
