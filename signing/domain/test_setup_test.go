package domain

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

	Init(&Config{
		MaxNumOfEmployeeManager: 5,
		DefaultGracePeriodDays:  30,
		VerificationCodeExpiry:  300,
		AccessTokenExpiry:       3600,
		MaxNumOfFailedLogin:     5,
		IntervalOfCreatingVC:    60,
		CommunityManagerLinkId:  "fake_link",
		SourceOfCLAPDF:          []string{"https://gitee.com"},
		MaxSizeOfCLAContent:     2 << 20,
		FileTypeOfCLAContent:    "pdf",
	})

	os.Exit(m.Run())
}
