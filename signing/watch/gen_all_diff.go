package watch

import (
	"fmt"
	"sort"
	"time"

	"github.com/beego/beego/v2/core/logs"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/message"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/util"
)

type repoLink interface {
	ListAll() ([]repository.LinkCLA, error)
}

func (impl *claUpdatedWatchImpl) genAllDiffFile() {
	interval := impl.config.genAllDiffInterval()
	timer := time.NewTimer(interval)
	for {
		select {
		case <-impl.stop:
			timer.Stop()
			impl.wg.Done()
			return
		case <-timer.C:
			impl.handleJob()
			timer.Reset(interval)
		}
	}
}

func (impl *claUpdatedWatchImpl) handleJob() {
	links, err := impl.link.ListAll()
	if err != nil {
		logs.Error("list all link failed: ", err)
		return
	}

	for i := range links {
		classifyCLA := impl.classifyCLA(&links[i])
		for j := range classifyCLA {
			impl.handleClassifiedCLAs(links[i].Id, classifyCLA[j])
		}
	}
}

func (impl *claUpdatedWatchImpl) classifyCLA(link *repository.LinkCLA) map[string][]*domain.CLA {
	classify := make(map[string][]*domain.CLA)

	f := func(cla *domain.CLA) {
		key := fmt.Sprintf("%s_%s", cla.Type.CLAType(), cla.Language.Language())
		classify[key] = append(classify[key], cla)
	}

	for i := range link.Clas {
		f(&link.Clas[i])
	}

	for i := range link.RemovedCLAs {
		f(&link.RemovedCLAs[i])
	}

	return classify
}

func (impl *claUpdatedWatchImpl) handleClassifiedCLAs(linkId string, clas []*domain.CLA) {
	sort.Slice(clas, func(i, j int) bool {
		return clas[i].Id > clas[j].Id
	})

	for i := 0; i < len(clas)-1; i++ {
		others := clas[i+1:]
		for j := range others {
			impl.sendGenDiffEvent(linkId, others[j].Id, clas[i].Id)
		}
	}
}

func (impl *claUpdatedWatchImpl) sendGenDiffEvent(linkId, oldCLAId, newCLAId string) {
	diffFile := impl.localCLA.LocalPathOfDiff(
		&domain.CLAIndex{
			LinkId: linkId,
			CLAId:  newCLAId,
		},
		oldCLAId,
	)

	if !util.IsFileNotExist(diffFile) {
		return
	}

	impl.genCLADiffByCron <- message.CLAUpdatedMsg{
		LinkId:   linkId,
		OldCLAId: oldCLAId,
		NewCLAId: newCLAId,
	}
}
