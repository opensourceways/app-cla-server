package messageimpl

import (
	kafka "github.com/opensourceways/kafka-lib/agent"
)

type Config struct {
	kafka.Config

	Topics Topics `json:"topics"  required:"true"`
}

type Topics struct {
	NewSignedCorpCLA string `json:"new_signed_corp_cla" required:"true"`
}
