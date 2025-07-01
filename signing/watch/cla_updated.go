package watch

import (
	"os/exec"
	"sync"

	"github.com/beego/beego/v2/core/logs"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/message"
	"github.com/opensourceways/app-cla-server/util"
)

var claUpdatedWatchInstance *claUpdatedWatchImpl

func CLAUpdatedWatchStart(lk repoLink, lc localCLA, cfg *CLAUpdateConfig, py string) {
	claUpdatedWatchInstance = &claUpdatedWatchImpl{
		config:           cfg,
		link:             lk,
		localCLA:         lc,
		pythonBin:        py,
		stop:             make(chan struct{}),
		genCLADiff:       make(chan message.CLAUpdatedMsg, cfg.MsgChannelSize),
		genCLADiffByCron: make(chan message.CLAUpdatedMsg, cfg.MsgChannelSize),
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
	select {
	case claUpdatedWatchInstance.genCLADiff <- msg:
	default:
		logs.Error("send cla updated event failed: ", msg)
	}
}

type localCLA interface {
	LocalPath(*domain.CLAIndex) string
	LocalPathOfDiff(index *domain.CLAIndex, signedClaId string) string
}

type claUpdatedWatchImpl struct {
	config *CLAUpdateConfig

	link      repoLink
	localCLA  localCLA
	pythonBin string

	wg               sync.WaitGroup
	stop             chan struct{}
	genCLADiff       chan message.CLAUpdatedMsg
	genCLADiffByCron chan message.CLAUpdatedMsg
}

func (impl *claUpdatedWatchImpl) start() {
	impl.wg.Add(1)
	go impl.subscribeGenCLADiff()

	impl.wg.Add(1)
	go impl.genAllDiffFile()
}

func (impl *claUpdatedWatchImpl) exit() {
	close(impl.stop)

	impl.wg.Wait()
}

func (impl *claUpdatedWatchImpl) subscribeGenCLADiff() {
	for {
		select {
		case <-impl.stop:
			impl.wg.Done()
			return
		case msgPrimary := <-impl.genCLADiff:
			impl.handleGenCLADiff(msgPrimary)
		case msgSecondary := <-impl.genCLADiffByCron:
			impl.handlePrimaryAgain()
			impl.handleGenCLADiff(msgSecondary)
		}
	}
}

// handlePrimaryAgain This is done to prioritize the processing of genCLADiff.
// Using a for loop is to prevent genCLADiffByCron tasks from being inserted into consecutive genCLADiff tasks.
// Only after genCLADiff is completed can genCLADiffByCron be executed.
func (impl *claUpdatedWatchImpl) handlePrimaryAgain() {
	for {
		select {
		case msgPrimary := <-impl.genCLADiff:
			impl.handleGenCLADiff(msgPrimary)
		default:
			return
		}
	}
}

func (impl *claUpdatedWatchImpl) handleGenCLADiff(msg message.CLAUpdatedMsg) {
	diffFile := impl.localCLA.LocalPathOfDiff(
		&domain.CLAIndex{
			LinkId: msg.LinkId,
			CLAId:  msg.NewCLAId,
		},
		msg.OldCLAId,
	)

	if !util.IsFileNotExist(diffFile) {
		return
	}

	oldPDFPath := impl.localCLA.LocalPath(&domain.CLAIndex{
		LinkId: msg.LinkId,
		CLAId:  msg.OldCLAId,
	})

	newPDFPath := impl.localCLA.LocalPath(&domain.CLAIndex{
		LinkId: msg.LinkId,
		CLAId:  msg.NewCLAId,
	})

	cmd := exec.Command(impl.pythonBin, "./util/generate_diff.py", oldPDFPath, newPDFPath, diffFile)
	for i := 0; i < impl.config.PythonRetryTimes; i++ {
		out, err := cmd.Output()
		if err == nil {
			break
		}

		logs.Error("gen pdf diff failed: ", diffFile, string(out), err)
	}
}
