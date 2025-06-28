package watch

import (
	"os/exec"
	"sync"

	"github.com/beego/beego/v2/core/logs"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/message"
)

var claUpdatedWatchInstance *claUpdatedWatchImpl

func CLAUpdatedWatchStart(lc localCLA, cfg *Config, py string) {
	claUpdatedWatchInstance = &claUpdatedWatchImpl{
		config:     cfg,
		localCLA:   lc,
		pythonBin:  py,
		genCLADiff: make(chan message.CLAUpdatedMsg, cfg.MsgChannelSize),
		sendEmail:  make(chan message.CLAUpdatedMsg, cfg.MsgChannelSize),
	}

	claUpdatedWatchInstance.start()
}

func CLAUpdateWatchStop() {
	if claUpdatedWatchInstance != nil {
		claUpdatedWatchInstance.exit()

		logs.Info("stop watching cla updated")
	}
}

func SendCLAUpdatedEvent(msg message.CLAUpdatedMsg) {
	if claUpdatedWatchInstance.needStop {
		return
	}

	claUpdatedWatchInstance.genCLADiff <- msg
	claUpdatedWatchInstance.sendEmail <- msg
}

type localCLA interface {
	LocalPath(*domain.CLAIndex) string
	LocalPathOfDiff(index *domain.CLAIndex, signedClaId string) string
}

type claUpdatedWatchImpl struct {
	wg       sync.WaitGroup
	needStop bool

	config     *Config
	localCLA   localCLA
	pythonBin  string
	genCLADiff chan message.CLAUpdatedMsg
	sendEmail  chan message.CLAUpdatedMsg
}

func (impl *claUpdatedWatchImpl) start() {
	go impl.subscribeGenCLADiff()
	go impl.subscribeSendEmail()
}

func (impl *claUpdatedWatchImpl) exit() {
	impl.needStop = true

	close(impl.genCLADiff)
	close(impl.sendEmail)

	impl.wg.Wait()
}

func (impl *claUpdatedWatchImpl) subscribeGenCLADiff() {
	impl.wg.Add(1)

	for v := range impl.genCLADiff {
		impl.handleGenCLADiff(v)
	}

	impl.wg.Done()
}

func (impl *claUpdatedWatchImpl) subscribeSendEmail() {
	impl.wg.Add(1)

	// handle

	impl.wg.Done()
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
	},
		msg.OldCLAId,
	)

	for i := 0; i < impl.config.PythonRetryTimes; i++ {
		cmd := exec.Command(impl.pythonBin, "./util/generate_diff.py", oldPDFPath, newPDFPath, diffFile)
		if out, err := cmd.Output(); err != nil {
			logs.Error("gen pdf diff failed: ", diffFile, string(out), err)
		} else {
			return
		}
	}
}
