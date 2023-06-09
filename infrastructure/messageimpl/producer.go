package messageimpl

import (
	kafka "github.com/opensourceways/kafka-lib/agent"
)

var producerInstance *producer

func Producer() *producer {
	return producerInstance
}

type producer struct {
	topics Topics
}

func (p *producer) NotifyNewSignedCorpCLA(e *NewSignedCorpCLA) error {
	return send(p.topics.NewSignedCorpCLA, e)
}

type eventMessage interface {
	message() ([]byte, error)
}

func send(topic string, v eventMessage) error {
	body, err := v.message()
	if err != nil {
		return err
	}

	return kafka.Publish(topic, body)
}
