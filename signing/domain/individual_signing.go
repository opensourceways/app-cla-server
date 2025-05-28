package domain

import "github.com/opensourceways/app-cla-server/util"

const (
	individualSigningActionSign    = "sign"
	individualSigningActionConfirm = "confirm"
)

type IndividualSigning struct {
	Id      string
	Link    LinkInfo
	Rep     Representative
	Date    string
	AllInfo AllSingingInfo
	Version int
	Logs    []IndividualSigningLog
}

type IndividualSigningLog struct {
	Date   string
	ClaId  string
	Action string
}

func (i *IndividualSigning) UpdateClaId(claId string) {
	i.Link.CLAId = claId
	i.AddConfirmLog(util.Date(), claId)
}

func (i *IndividualSigning) AddSignLog(date, claId string) {
	i.Logs = []IndividualSigningLog{
		{
			Date:   date,
			ClaId:  claId,
			Action: individualSigningActionSign,
		},
	}
}

func (i *IndividualSigning) AddConfirmLog(date, claId string) {
	if len(i.Logs) == 0 {
		i.AddSignLog(i.Date, i.Link.CLAId)
	}

	i.Logs = append(i.Logs, IndividualSigningLog{
		Date:   date,
		ClaId:  claId,
		Action: individualSigningActionConfirm,
	})
}
