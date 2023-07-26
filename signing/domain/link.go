package domain

import (
	"fmt"
	"strconv"
	"time"

	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

type EmailInfo struct {
	Addr     dp.EmailAddr
	Platform string
}

type OrgInfo struct {
	Platform string
	Org      string
	Alias    string
}

func (v *OrgInfo) LinkId() string {
	return fmt.Sprintf("%s_%s-%d", v.Platform, v.Org, time.Now().UnixNano())
}

type Link struct {
	Id        string
	Org       OrgInfo
	Email     EmailInfo
	CLAs      []CLA
	Submitter string
	CLANum    int
	Version   int
}

func (link *Link) AddCLA(cla *CLA) error {
	if _, ok := link.posOfCLA(cla); ok {
		return NewDomainError(ErrorCodeCLAExists)
	}

	cla.Id = strconv.Itoa(link.CLANum)
	link.CLANum += 1

	return nil
}

func (link *Link) FindCLA(index string) *CLA {
	for i := range link.CLAs {
		if link.CLAs[i].Id == index {
			return &link.CLAs[i]
		}
	}

	return nil
}

func (link *Link) posOfCLA(cla *CLA) (int, bool) {
	for i := range link.CLAs {
		if link.CLAs[i].isMe(cla) {
			return i, true
		}
	}

	return 0, false
}
