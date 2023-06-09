package messageimpl

import (
	kafka "github.com/opensourceways/kafka-lib/agent"
	"github.com/sirupsen/logrus"
)

func Init(cfg *Config, log *logrus.Entry) error {
	producerInstance = &producer{cfg.Topics}

	return kafka.Init(&cfg.Config, log)
}

func Exit() {
	kafka.Exit()
}
