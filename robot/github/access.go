package github

import "github.com/opensourceways/app-cla-server/robot/github/webhook"

type access struct {
	getHamc func() []byte
}

func (a access) checkWebhook(payload []byte, getHeader func(string) string) (string, string, int, error) {
	return webhook.ValidateWebhook(getHeader, payload, a.getHamc)
}
