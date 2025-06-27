package watch

import (
	"os/exec"

	"github.com/beego/beego/v2/core/logs"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/message"
)

var claUpdatedWatchInstance *claUpdatedWatchImpl

func CLAUpdatedWatchStart(lc localCLA, py string) {
	claUpdatedWatchInstance = &claUpdatedWatchImpl{
		localCLA:   lc,
		pythonBin:  py,
		genCLADiff: make(chan message.CLAUpdatedMsg, 10),
		sendEmail:  make(chan message.CLAUpdatedMsg, 10),
	}

	claUpdatedWatchInstance.start()
}

func ClaUpdatedWatchInstance() *claUpdatedWatchImpl {
	return claUpdatedWatchInstance
}

type localCLA interface {
	LocalPath(*domain.CLAIndex) string
	LocalPathOfDiff(index *domain.CLAIndex, signedClaId string) string
}

type claUpdatedWatchImpl struct {
	localCLA   localCLA
	pythonBin  string
	genCLADiff chan message.CLAUpdatedMsg
	sendEmail  chan message.CLAUpdatedMsg
}

func (impl *claUpdatedWatchImpl) Send(msg message.CLAUpdatedMsg) {
	impl.genCLADiff <- msg
	impl.sendEmail <- msg
}

func (impl *claUpdatedWatchImpl) start() {
	go impl.subscribeGenCLADiff()
	go impl.subscribeSendEmail()
}

func (impl *claUpdatedWatchImpl) subscribeGenCLADiff() {
	for v := range impl.genCLADiff {
		impl.handleGenCLADiff(v)
	}
}

func (impl *claUpdatedWatchImpl) subscribeSendEmail() {

}

func (impl *claUpdatedWatchImpl) handleGenCLADiff(msg message.CLAUpdatedMsg) {
	oldPDFPath := impl.localCLA.LocalPath(&domain.CLAIndex{
		LinkId: msg.LinkId,
		CLAId:  msg.OldCLAId,
	})

	newPDFPath := impl.localCLA.LocalPath(&domain.CLAIndex{
		LinkId: msg.LinkId,
		CLAId:  msg.NewCLAId,
	})

	diffFile := impl.localCLA.LocalPathOfDiff(&domain.CLAIndex{
		LinkId: msg.LinkId,
		CLAId:  msg.NewCLAId,
	}, msg.OldCLAId,
	)

	cmd := exec.Command(impl.pythonBin, "./util/generate_diff.py", oldPDFPath, newPDFPath, diffFile)
	if out, err := cmd.Output(); err != nil {
		logs.Error("gen pdf diff failed: ", diffFile, string(out), err)
	}
}
