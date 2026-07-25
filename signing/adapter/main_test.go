package adapter

import (
	"os"
	"testing"

	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func TestMain(m *testing.M) {
	dp.Init(&dp.Config{
		MaxLengthOfName:     100,
		MaxLengthOfTitle:    200,
		MaxLengthOfEmail:    200,
		MaxLengthOfAccount:  100,
		MaxLengthOfCorpName: 200,
	})

	os.Exit(m.Run())
}
